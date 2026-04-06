package filters

import (
	"strings"
	"testing"
)

func TestFilterPipenvEmpty(t *testing.T) {
	got, err := filterPipenv("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected passthrough for empty input, got %q", got)
	}
}

func TestFilterPipenvShort(t *testing.T) {
	raw := `Pipfile: /home/user/myproject/Pipfile
Using /usr/bin/python3.11 (3.11.0) to create virtualenv...
✔ Successfully created virtual environment!
Virtualenv location: /home/user/.local/share/virtualenvs/myproject-abc123
Installing dependencies from Pipfile.lock...
✔ Success!`

	got, err := filterPipenv(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Short output should not expand
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	if filteredTokens > rawTokens {
		t.Errorf("filter expanded output: raw=%d tokens, filtered=%d tokens", rawTokens, filteredTokens)
	}
}

func TestFilterPipenvInstall(t *testing.T) {
	raw := `Creating a virtualenv for this project...
Pipfile: /home/user/myproject/Pipfile
Using /usr/bin/python3.11 (3.11.0) to create virtualenv...
✔ Successfully created virtual environment!
Virtualenv location: /home/user/.local/share/virtualenvs/myproject-abc123

Installing dependencies from Pipfile.lock...
Pipfile.lock (abc123) out of date, updating to (def456)...
Running $ pipenv lock then $ pipenv sync.
Locking [packages] dependencies...
Building requirements...
Resolving dependencies...
✔ Success!
Updated Pipfile.lock (def456)!
Locking [dev-packages] dependencies...
Building requirements...
Resolving dependencies...
✔ Success!
Updated Pipfile.lock (def456)!
Installing dependencies...
  Installing requests (2.31.0)
  Installing certifi (2023.7.22)
  Installing charset-normalizer (3.3.0)
  Installing urllib3 (2.0.4)
  Installing idna (3.4)
  Installing Flask (2.3.2)
  Installing Werkzeug (2.3.6)
  Installing click (8.1.6)
  Installing Jinja2 (3.1.2)
  Installing MarkupSafe (2.1.3)
  Installing itsdangerous (2.1.2)
To activate this project's virtualenv, run pipenv shell.
Alternatively, run a command inside the virtualenv with pipenv run.`

	got, err := filterPipenv(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show virtualenv location
	if !strings.Contains(got, "Virtualenv location:") {
		t.Errorf("expected virtualenv location in output, got:\n%s", got)
	}
	if !strings.Contains(got, "myproject-abc123") {
		t.Errorf("expected virtualenv path in output, got:\n%s", got)
	}
	// Should show installing deps line
	if !strings.Contains(got, "Installing dependencies from Pipfile.lock") {
		t.Errorf("expected installing deps line, got:\n%s", got)
	}
	// Should show success status
	if !strings.Contains(got, "Success") {
		t.Errorf("expected success status in output, got:\n%s", got)
	}

	// Token savings >= 50%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 50.0 {
		t.Errorf("expected >=50%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterPipenvError(t *testing.T) {
	raw := `Creating a virtualenv for this project...
Pipfile: /home/user/myproject/Pipfile
Using /usr/bin/python3.11 (3.11.0) to create virtualenv...
✔ Successfully created virtual environment!
Virtualenv location: /home/user/.local/share/virtualenvs/myproject-abc123

Installing dependencies from Pipfile.lock...
ERROR: No matching distribution found for somepackage==99.99.99
Error: Aborting due to failure while installing packages`

	got, err := filterPipenv(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain error line
	if !strings.Contains(got, "ERROR") {
		t.Errorf("expected ERROR in output, got:\n%s", got)
	}
	if !strings.Contains(got, "somepackage") {
		t.Errorf("expected error details in output, got:\n%s", got)
	}
}

func TestFilterPipenvSanityCheck(t *testing.T) {
	raw := `Creating a virtualenv for this project...
Pipfile: /home/user/myproject/Pipfile
Using /usr/bin/python3.11 (3.11.0) to create virtualenv...
✔ Successfully created virtual environment!
Virtualenv location: /home/user/.local/share/virtualenvs/myproject-abc123

Installing dependencies from Pipfile.lock...
Locking [packages] dependencies...
Building requirements...
Resolving dependencies...
✔ Success!
Updated Pipfile.lock!
Locking [dev-packages] dependencies...
Building requirements...
Resolving dependencies...
✔ Success!
Updated Pipfile.lock!
To activate this project's virtualenv, run pipenv shell.
Alternatively, run a command inside the virtualenv with pipenv run.`

	got, err := filterPipenv(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) > len(raw) {
		t.Errorf("filter expanded output: raw=%d bytes, filtered=%d bytes", len(raw), len(got))
	}
}
