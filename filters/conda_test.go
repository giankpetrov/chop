package filters

import (
	"fmt"
	"strings"
	"testing"
)

func TestFilterCondaInstallEmpty(t *testing.T) {
	got, err := filterCondaInstall("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected passthrough for empty input, got %q", got)
	}
}

func TestFilterCondaInstallAlreadySatisfied(t *testing.T) {
	raw := `Collecting package metadata (current_repodata.json): done
Solving environment: done
All requested packages already installed.`

	got, err := filterCondaInstall(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should not expand
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	if filteredTokens > rawTokens {
		t.Errorf("filter expanded output: raw=%d tokens, filtered=%d tokens", rawTokens, filteredTokens)
	}
}

func TestFilterCondaInstallLarge(t *testing.T) {
	raw := `Collecting package metadata (current_repodata.json): done
Solving environment: done

## Package Plan ##

  environment location: /home/user/miniconda3/envs/myenv

  added / updated specs:
    - numpy
    - pandas
    - scipy
    - matplotlib

The following packages will be downloaded:

    package                    |            build
    ---------------------------|-----------------
    numpy-1.24.0               |   py311h7f486d5_0    15.2 MB
    pandas-2.0.0               |   py311h7f486d5_0    12.1 MB
    scipy-1.11.0               |   py311h7f486d5_0    22.4 MB
    matplotlib-3.7.1           |   py311h7f486d5_0     7.8 MB
    pillow-9.5.0               |   py311hcec6c5f_0     3.1 MB
    kiwisolver-1.4.4           |   py311h6a678d5_0     0.1 MB
    pyparsing-3.0.9            |   py311h06a4308_0     0.1 MB
    python-dateutil-2.8.2      |             pyhd3eb1b0_0     0.2 MB
    pytz-2023.3                |   py311h06a4308_0     0.3 MB
    six-1.16.0                 |   py311h06a4308_0     0.0 MB
                                           ----
                                           Total: 61.3 MB

The following NEW packages will be INSTALLED:
  kiwisolver      pkgs/main/linux-64::kiwisolver-1.4.4-py311h6a678d5_0
  matplotlib      pkgs/main/linux-64::matplotlib-3.7.1-py311h7f486d5_0
  numpy           pkgs/main/linux-64::numpy-1.24.0-py311h7f486d5_0
  pandas          pkgs/main/linux-64::pandas-2.0.0-py311h7f486d5_0
  pillow          pkgs/main/linux-64::pillow-9.5.0-py311hcec6c5f_0
  pyparsing       pkgs/main/linux-64::pyparsing-3.0.9-py311h06a4308_0
  python-dateutil pkgs/main/linux-64::python-dateutil-2.8.2-pyhd3eb1b0_0
  pytz            pkgs/main/linux-64::pytz-2023.3-py311h06a4308_0
  scipy           pkgs/main/linux-64::scipy-1.11.0-py311h7f486d5_0
  six             pkgs/main/linux-64::six-1.16.0-py311h06a4308_0

Downloading and Extracting Packages:
numpy-1.24.0         | 15.2 MB   | ############################################ | 100%
pandas-2.0.0         | 12.1 MB   | ############################################ | 100%
scipy-1.11.0         | 22.4 MB   | ############################################ | 100%
matplotlib-3.7.1     |  7.8 MB   | ############################################ | 100%
pillow-9.5.0         |  3.1 MB   | ############################################ | 100%
kiwisolver-1.4.4     |  0.1 MB   | ############################################ | 100%
pyparsing-3.0.9      |  0.1 MB   | ############################################ | 100%
python-dateutil-2.8.2 |  0.2 MB  | ############################################ | 100%
pytz-2023.3          |  0.3 MB   | ############################################ | 100%
six-1.16.0           |  0.0 MB   | ############################################ | 100%
Preparing transaction: done
Verifying transaction: done
Executing transaction: done`

	got, err := filterCondaInstall(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show specs
	if !strings.Contains(got, "numpy") {
		t.Errorf("expected specs in output, got:\n%s", got)
	}
	// Should show total download size
	if !strings.Contains(got, "61.3 MB") {
		t.Errorf("expected total download size, got:\n%s", got)
	}
	// Should show new package count
	if !strings.Contains(got, "new packages installed") {
		t.Errorf("expected new package count, got:\n%s", got)
	}
	// Should show final status
	if !strings.Contains(got, "transaction") {
		t.Errorf("expected transaction status, got:\n%s", got)
	}

	// Token savings >= 50%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 50.0 {
		t.Errorf("expected >=50%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterCondaListFew(t *testing.T) {
	raw := `# packages in environment at /home/user/miniconda3/envs/myenv:
#
# Name                    Version                   Build  Channel
certifi                   2023.7.22          py311h06a4308_0
numpy                     1.24.0             py311h7f486d5_0
pandas                    2.0.0              py311h7f486d5_0
pip                       23.2.1             py311h06a4308_0
python                    3.11.4             h955ad1f_0
setuptools                68.0.0             py311h06a4308_0
six                       1.16.0             py311h06a4308_0
wheel                     0.38.4             py311h06a4308_0`

	got, err := filterCondaList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 8 packages is <= 15, should passthrough
	if got != raw {
		t.Errorf("expected passthrough for few packages, got:\n%s", got)
	}
}

func TestFilterCondaListMany(t *testing.T) {
	var lines []string
	lines = append(lines, "# packages in environment at /home/user/miniconda3/envs/myenv:")
	lines = append(lines, "#")
	lines = append(lines, "# Name                    Version                   Build  Channel")
	pkgs := []string{
		"attrs", "beautifulsoup4", "certifi", "cffi", "charset-normalizer",
		"click", "colorama", "cryptography", "filelock", "idna",
		"importlib-metadata", "iniconfig", "jinja2", "markupsafe", "numpy",
		"packaging", "pandas", "pillow", "pip", "pluggy",
		"pycparser", "pyparsing", "pytest", "python", "python-dateutil",
		"pytz", "requests", "scipy", "setuptools", "six",
		"tomli", "typing-extensions", "urllib3", "wheel", "zipp",
	}
	for _, pkg := range pkgs {
		lines = append(lines, fmt.Sprintf("%-25s 1.0.0              py311h06a4308_0", pkg))
	}
	raw := strings.Join(lines, "\n")

	got, err := filterCondaList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 35 packages > 15, should compress
	if got == raw {
		t.Errorf("expected compression for 35 packages, got passthrough")
	}
	// Should contain count
	if !strings.Contains(got, "35") {
		t.Errorf("expected package count (35) in output, got:\n%s", got)
	}
}

func TestFilterCondaSanityCheck(t *testing.T) {
	// Install output sanity
	raw := `Collecting package metadata (current_repodata.json): done
Solving environment: done

## Package Plan ##

  environment location: /home/user/miniconda3/envs/myenv

  added / updated specs:
    - numpy

The following NEW packages will be INSTALLED:
  numpy  pkgs/main/linux-64::numpy-1.24.0-py311h7f486d5_0

Preparing transaction: done
Verifying transaction: done
Executing transaction: done`

	got, err := filterCondaInstall(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) > len(raw) {
		t.Errorf("install filter expanded output: raw=%d bytes, filtered=%d bytes", len(raw), len(got))
	}

	// List output sanity
	var listLines []string
	listLines = append(listLines, "# packages in environment at /home/user/miniconda3/envs/myenv:")
	listLines = append(listLines, "#")
	listLines = append(listLines, "# Name                    Version                   Build  Channel")
	for i := 0; i < 25; i++ {
		listLines = append(listLines, fmt.Sprintf("package%-3d                1.0.%d              py311h06a4308_0", i, i))
	}
	rawList := strings.Join(listLines, "\n")

	gotList, err := filterCondaList(rawList)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gotList) > len(rawList) {
		t.Errorf("list filter expanded output: raw=%d bytes, filtered=%d bytes", len(rawList), len(gotList))
	}
}
