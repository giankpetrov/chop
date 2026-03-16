package filters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/giankpetrov/openchop/config"
)

func TestBuildExecFilter_Secure(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "pwned")
	// Malicious command that would create a file if passed to a shell
	execCmd := "echo hello; touch " + tmpFile

	cf := &config.CustomFilter{
		Exec:    execCmd,
		Trusted: true,
	}

	fn := BuildUserFilter(cf)
	_, err := fn("some input")
	// It should FAIL to execute because "echo hello; touch ..." is not a valid executable name
	// OR it should succeed but NOT create the file because ";" is treated as a literal argument
	// In our implementation, parts[0] is "echo" and the rest are arguments.
	// So it should run `echo "hello;" "touch" "/tmp/..."`

	if err != nil {
		t.Fatalf("filter failed: %v", err)
	}

	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Errorf("vulnerability STILL PRESENT: %s was created", tmpFile)
	}
}

func TestBuildUserFilter_UntrustedExec(t *testing.T) {
	cf := &config.CustomFilter{
		Exec:    "ls",
		Trusted: false,
	}

	fn := BuildUserFilter(cf)
	if fn != nil {
		t.Error("expected nil for untrusted exec filter")
	}
}

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{`ls -la`, []string{"ls", "-la"}},
		{`echo "hello world"`, []string{"echo", "hello world"}},
		{`echo 'hello world'`, []string{"echo", "hello world"}},
		{`python3 -c "print('hi')"`, []string{"python3", "-c", "print('hi')"}},
		{`  trimmed   `, []string{"trimmed"}},
		{`cmd "arg with spaces" 'another one'`, []string{"cmd", "arg with spaces", "another one"}},
	}

	for _, tc := range tests {
		got := splitCommand(tc.input)
		if len(got) != len(tc.expected) {
			t.Errorf("input %q: expected %v, got %v", tc.input, tc.expected, got)
			continue
		}
		for i := range got {
			if got[i] != tc.expected[i] {
				t.Errorf("input %q: expected %v, got %v", tc.input, tc.expected, got)
				break
			}
		}
	}
}
