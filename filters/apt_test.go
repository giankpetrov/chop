package filters

import (
	"strings"
	"testing"
)

var aptFixture = `Reading package lists... Done
Building dependency tree... Done
Reading state information... Done
The following NEW packages will be installed:
  curl libcurl4
0 upgraded, 2 newly installed, 0 to remove and 12 not upgraded.
Need to get 345 kB of archives.
After this operation, 1,023 kB of additional disk space will be used.
Get:1 http://archive.ubuntu.com/ubuntu focal/main amd64 libcurl4 amd64 7.68.0-1ubuntu2 [233 kB]
Get:2 http://archive.ubuntu.com/ubuntu focal/main amd64 curl amd64 7.68.0-1ubuntu2 [161 kB]
Fetched 394 kB in 1s (394 kB/s)
Selecting previously unselected package libcurl4:amd64.
(Reading database ... 12345 files and directories currently installed.)
Preparing to unpack .../libcurl4_7.68.0-1ubuntu2_amd64.deb ...
Unpacking libcurl4:amd64 (7.68.0-1ubuntu2) ...
Selecting previously unselected package curl.
Preparing to unpack .../curl_7.68.0-1ubuntu2_amd64.deb ...
Unpacking curl (7.68.0-1ubuntu2) ...
Setting up libcurl4:amd64 (7.68.0-1ubuntu2) ...
Setting up curl (7.68.0-1ubuntu2) ...
Processing triggers for man-db (2.9.1-1) ...
`

func TestFilterApt(t *testing.T) {
	got, err := filterApt(aptFixture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should keep "Setting up" lines
	if !strings.Contains(got, "Setting up libcurl4") {
		t.Errorf("expected 'Setting up libcurl4' preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "Setting up curl") {
		t.Errorf("expected 'Setting up curl' preserved, got:\n%s", got)
	}

	// Should keep summary line
	if !strings.Contains(got, "newly installed") {
		t.Errorf("expected summary line preserved, got:\n%s", got)
	}

	// Should drop "Get:" lines
	if strings.Contains(got, "Get:") {
		t.Errorf("expected 'Get:' lines dropped, got:\n%s", got)
	}

	// Should drop "Fetched" lines
	if strings.Contains(got, "Fetched") {
		t.Errorf("expected 'Fetched' lines dropped, got:\n%s", got)
	}

	// Should drop "Unpacking" lines
	if strings.Contains(got, "Unpacking") {
		t.Errorf("expected 'Unpacking' lines dropped, got:\n%s", got)
	}

	// Should drop "Selecting previously" lines
	if strings.Contains(got, "Selecting previously") {
		t.Errorf("expected 'Selecting previously' lines dropped, got:\n%s", got)
	}
}

func TestFilterAptEmpty(t *testing.T) {
	got, err := filterApt("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestAptRouted(t *testing.T) {
	for _, sub := range []string{"install", "upgrade", "update", "dist-upgrade"} {
		if get("apt", []string{sub}) == nil {
			t.Errorf("expected non-nil filter for apt %s", sub)
		}
		if get("apt-get", []string{sub}) == nil {
			t.Errorf("expected non-nil filter for apt-get %s", sub)
		}
	}
}

func TestAptRoutedNilForOther(t *testing.T) {
	if get("apt", []string{"show"}) != nil {
		t.Error("expected nil filter for apt show")
	}
}
