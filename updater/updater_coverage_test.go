package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// --- dataDir ---

func TestDataDir_CreatesDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir, err := dataDir()
	if err != nil {
		t.Fatalf("dataDir() error: %v", err)
	}

	want := filepath.Join(home, ".local", "share", "chop")
	if dir != want {
		t.Errorf("dataDir() = %q, want %q", dir, want)
	}

	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Errorf("dataDir() did not create directory at %q", dir)
	}
}

// --- lastCheckPath / pendingUpdatePath ---

func TestLastCheckPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	p, err := lastCheckPath()
	if err != nil {
		t.Fatalf("lastCheckPath() error: %v", err)
	}
	if !strings.HasSuffix(p, "last-update-check") {
		t.Errorf("lastCheckPath() = %q, want suffix 'last-update-check'", p)
	}
}

func TestPendingUpdatePath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	p, err := pendingUpdatePath()
	if err != nil {
		t.Fatalf("pendingUpdatePath() error: %v", err)
	}
	if !strings.HasSuffix(p, "pending-update") {
		t.Errorf("pendingUpdatePath() = %q, want suffix 'pending-update'", p)
	}
}

// --- clearUpdateAvailable / recordUpdateAvailable ---

func TestRecordAndClearUpdateAvailable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// Record a version
	recordUpdateAvailable("v9.9.9")

	p, err := updateAvailablePath()
	if err != nil {
		t.Fatalf("updateAvailablePath() error: %v", err)
	}

	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("expected file to exist after recordUpdateAvailable: %v", err)
	}
	if string(data) != "v9.9.9" {
		t.Errorf("recorded %q, want %q", string(data), "v9.9.9")
	}

	// Clear it
	clearUpdateAvailable()
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Error("clearUpdateAvailable() did not remove the file")
	}
}

func TestClearUpdateAvailable_NoFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// Should not panic or error when file does not exist
	clearUpdateAvailable()
}

func TestRecordUpdateAvailable_Overwrite(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	recordUpdateAvailable("v1.0.0")
	recordUpdateAvailable("v2.0.0")

	p, _ := updateAvailablePath()
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("os.ReadFile: %v", err)
	}
	if string(data) != "v2.0.0" {
		t.Errorf("got %q, want %q", string(data), "v2.0.0")
	}
}

// --- fetchReleaseFile (via httptest) ---

func TestFetchReleaseFile_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake-content"))
	}))
	defer srv.Close()

	// fetchReleaseFile builds a URL using the repo constant and GitHub base URL.
	// We can't inject the test server into fetchReleaseFile directly without
	// refactoring, but we can test it by overriding httpClient temporarily.
	origClient := httpClient
	httpClient = srv.Client()
	defer func() { httpClient = origClient }()

	// Since fetchReleaseFile always hits github.com, test the lower-level
	// download helper that shares the same httpClient.
	dest := filepath.Join(t.TempDir(), "out")
	// Serve content large enough to pass the size check
	payload := fakeBinaryPayload()
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(payload)
	}))
	defer srv2.Close()

	if err := download(srv2.URL, dest); err != nil {
		t.Fatalf("download via httptest server failed: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(payload) {
		t.Error("downloaded content does not match payload")
	}
}

func TestFetchReleaseFile_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer srv.Close()

	origClient := httpClient
	httpClient = &http.Client{}
	defer func() { httpClient = origClient }()

	// Reach fetchReleaseFile logic via download (same HTTP path)
	dest := filepath.Join(t.TempDir(), "out")
	err := download(srv.URL, dest)
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 in error, got: %v", err)
	}
}

// --- hashFile ---

func TestHashFile_KnownContent(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "input")
	content := []byte("chop test content")
	os.WriteFile(f, content, 0o644)

	h := sha256.Sum256(content)
	want := hex.EncodeToString(h[:])

	got, err := hashFile(f)
	if err != nil {
		t.Fatalf("hashFile error: %v", err)
	}
	if got != want {
		t.Errorf("hashFile = %q, want %q", got, want)
	}
}

func TestHashFile_Missing(t *testing.T) {
	_, err := hashFile("/nonexistent/path/file")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// --- parseChecksum edge cases ---

func TestParseChecksum_EmptyLines(t *testing.T) {
	checksums := "\n\n\nabc123  chop-linux-amd64\n\n"
	got, err := parseChecksum(checksums, "chop-linux-amd64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "abc123" {
		t.Errorf("got %q, want %q", got, "abc123")
	}
}

func TestParseChecksum_MalformedLine(t *testing.T) {
	// Single-field lines (no space) should be skipped
	checksums := "justahash\nabc123  chop-linux-amd64\n"
	got, err := parseChecksum(checksums, "chop-linux-amd64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "abc123" {
		t.Errorf("got %q, want %q", got, "abc123")
	}
}

// --- buildBinaryName (additional coverage) ---

func TestBuildBinaryName_Format(t *testing.T) {
	name := buildBinaryName()
	// Must be chop-<goos>-<goarch>[.exe]
	parts := strings.SplitN(name, "-", 3)
	if len(parts) < 3 {
		t.Fatalf("buildBinaryName() = %q: expected at least 3 dash-separated parts", name)
	}
	if parts[0] != "chop" {
		t.Errorf("expected first segment 'chop', got %q", parts[0])
	}
	if parts[1] != runtime.GOOS {
		t.Errorf("expected GOOS %q, got %q", runtime.GOOS, parts[1])
	}
}

// --- IsDev edge cases ---

func TestIsDev_Dirty(t *testing.T) {
	if !IsDev("v1.2.3-dirty") {
		t.Error("IsDev(v1.2.3-dirty) should be true")
	}
	if !IsDev("dev") {
		t.Error("IsDev(dev) should be true")
	}
	if IsDev("v1.2.3") {
		t.Error("IsDev(v1.2.3) should be false")
	}
}

// --- replaceBinary edge cases ---

func TestReplaceBinary_DestMissing(t *testing.T) {
	// On non-Windows, os.Rename(src, dest) where dest doesn't exist is fine.
	// On Windows, the rename dance handles missing dest gracefully too.
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dest := filepath.Join(dir, "dest")

	os.WriteFile(src, []byte("new"), 0o700)
	// dest does not exist — should still succeed
	if err := replaceBinary(dest, src); err != nil {
		t.Fatalf("replaceBinary with no existing dest: %v", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Errorf("expected 'new', got %q", string(data))
	}
}

// --- ApplyPendingUpdate with auto-update off cleans up ---

func TestApplyPendingUpdate_AutoUpdateOff_CleansUp(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// Ensure auto-update is off
	SetAutoUpdate(false)

	// Write a fake pending marker with two parts (version + binary path)
	pendingPath, err := pendingUpdatePath()
	if err != nil {
		t.Fatal(err)
	}
	os.MkdirAll(filepath.Dir(pendingPath), 0o700)

	fakebin := filepath.Join(t.TempDir(), "chop.new")
	os.WriteFile(fakebin, []byte("binary"), 0o700)

	content := "v2.0.0\n" + fakebin
	os.WriteFile(pendingPath, []byte(content), 0o600)

	ApplyPendingUpdate("v1.0.0")

	// Marker should be gone
	if _, err := os.Stat(pendingPath); !os.IsNotExist(err) {
		t.Error("pending marker should be cleaned up when auto-update is off")
	}
	// Fake binary should also be gone
	if _, err := os.Stat(fakebin); !os.IsNotExist(err) {
		t.Error("downloaded binary should be cleaned up when auto-update is off")
	}
}
