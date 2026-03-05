package filters

import (
	"strings"
	"testing"
)

func TestFilterNgBuild(t *testing.T) {
	raw := `- Generating browser application bundles (phase 1)...
✔ Browser application bundle generation complete.
✔ Copying assets complete.
✔ Index html generation complete.

Initial Chunk Files           | Names         |  Raw Size | Estimated Transfer Size
main.js                       | main          | 250.00 kB |            65.00 kB
polyfills.js                  | polyfills     |  33.00 kB |            10.50 kB
styles.css                    | styles        |  75.00 kB |            10.00 kB

                              | Initial Total | 358.00 kB |            85.50 kB

Build at: 2024-01-15T10:30:00.000Z - Hash: abc123 - Time: 15000ms

Warning: some/file.ts depends on 'something'. CommonJS dependencies can cause optimization bailouts.`

	got, err := filterNgBuild(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "build ok") {
		t.Error("expected build ok")
	}
	if !strings.Contains(got, "1 warnings") {
		t.Error("expected warning count")
	}
}

func TestFilterNgTest(t *testing.T) {
	raw := `Chrome 120.0.0 (Mac): Executed 0 of 50 SUCCESS (0 secs / 0 secs)
Chrome 120.0.0 (Mac): Executed 50 of 50 (3 FAILED) (5.2 secs / 4.8 secs)

TOTAL: 3 FAILED, 47 SUCCESS`

	got, err := filterNgTest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "3 FAILED, 47 SUCCESS") {
		t.Errorf("expected summary, got: %s", got)
	}
}

func TestFilterNgServe(t *testing.T) {
	raw := `Initial Chunk Files   | Names         |  Raw Size
main.js               | main          | 250.00 kB
                      | Initial Total | 358.00 kB

Application bundle generation complete. [15.000 seconds]
** Angular Live Development Server is listening on localhost:4200, open your browser on http://localhost:4200/ **`

	got, err := filterNgServe(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "serving on") {
		t.Errorf("expected serving message, got: %s", got)
	}
}

func TestFilterNgBuild_Empty(t *testing.T) {
	got, err := filterNgBuild("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
