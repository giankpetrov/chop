package filters

import (
	"strings"
	"testing"
)

// --- Router dispatch tests ---

func TestGetDockerFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"ps"}, false},
		{[]string{"build"}, false},
		{[]string{"images"}, false},
		{[]string{"logs"}, false},
		{[]string{"rmi"}, false},
		{[]string{"inspect"}, false},
		{[]string{"stats"}, false},
		{[]string{"top"}, false},
		{[]string{"diff"}, false},
		{[]string{"history"}, false},
		{[]string{"network", "ls"}, false},
		{[]string{"network", "inspect"}, true},
		{[]string{"network"}, true},
		{[]string{"volume", "ls"}, false},
		{[]string{"volume", "prune"}, true},
		{[]string{"volume"}, true},
		{[]string{"system", "df"}, false},
		{[]string{"system", "prune"}, true},
		{[]string{"system"}, true},
		{[]string{"compose", "ps"}, false},
		{[]string{"compose", "unknown"}, true},
		{[]string{"unknown"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getDockerFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getDockerFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getDockerFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetDockerComposeFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{}, false}, // returns filterAutoDetect
		{[]string{"ps"}, false},
		{[]string{"build"}, false},
		{[]string{"logs"}, false},
		{[]string{"images"}, false},
		{[]string{"unknown"}, true},
	}
	for _, tc := range cases {
		got := getDockerComposeFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getDockerComposeFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getDockerComposeFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetGitFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"status"}, false},
		{[]string{"log"}, false},
		{[]string{"diff"}, false},
		{[]string{"show"}, false},
		{[]string{"branch"}, false},
		{[]string{"push"}, false},
		{[]string{"pull"}, false},
		{[]string{"fetch"}, false},
		{[]string{"remote"}, false},
		{[]string{"tag"}, false},
		{[]string{"checkout"}, false},
		{[]string{"reset"}, false},
		{[]string{"stash", "list"}, false},
		{[]string{"stash", "pop"}, true},
		{[]string{"stash"}, true},
		{[]string{"unknown"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getGitFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getGitFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getGitFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetKubectlFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"get"}, false},
		{[]string{"describe"}, false},
		{[]string{"logs"}, false},
		{[]string{"log"}, false},
		{[]string{"top"}, false},
		{[]string{"apply"}, false},
		{[]string{"delete"}, false},
		{[]string{"unknown"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getKubectlFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getKubectlFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getKubectlFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetHelmFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"install"}, false},
		{[]string{"upgrade"}, false},
		{[]string{"list"}, false},
		{[]string{"ls"}, false},
		{[]string{"status"}, false},
		{[]string{"unknown"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getHelmFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getHelmFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getHelmFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetTerraformFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"plan"}, false},
		{[]string{"apply"}, false},
		{[]string{"init"}, false},
		{[]string{"destroy"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getTerraformFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getTerraformFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getTerraformFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetNpmFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"install"}, false},
		{[]string{"i"}, false},
		{[]string{"update"}, false},
		{[]string{"up"}, false},
		{[]string{"upgrade"}, false},
		{[]string{"list"}, false},
		{[]string{"ls"}, false},
		{[]string{"view"}, false},
		{[]string{"info"}, false},
		{[]string{"show"}, false},
		{[]string{"test"}, false},
		{[]string{"t"}, false},
		{[]string{"run", "test"}, false},
		{[]string{"run", "build"}, false},
		{[]string{"run", "lint"}, false},
		{[]string{"run", "other"}, true},
		{[]string{"run"}, true},
		{[]string{"publish"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getNpmFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getNpmFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getNpmFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetGoFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"test"}, false},
		{[]string{"build"}, false},
		{[]string{"vet"}, false},
		{[]string{"run"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getGoFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getGoFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getGoFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetCargoFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"test"}, false},
		{[]string{"build"}, false},
		{[]string{"check"}, false},
		{[]string{"clippy"}, false},
		{[]string{"run"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getCargoFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getCargoFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getCargoFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetPnpmFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"install"}, false},
		{[]string{"i"}, false},
		{[]string{"add"}, false},
		{[]string{"list"}, false},
		{[]string{"ls"}, false},
		{[]string{"test"}, false},
		{[]string{"t"}, false},
		{[]string{"run"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getPnpmFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getPnpmFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getPnpmFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetYarnFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"install"}, false},
		{[]string{"add"}, false},
		{[]string{"list"}, false},
		{[]string{"test"}, false},
		{[]string{"run"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getYarnFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getYarnFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getYarnFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetBunFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"install"}, false},
		{[]string{"i"}, false},
		{[]string{"add"}, false},
		{[]string{"test"}, false},
		{[]string{"t"}, false},
		{[]string{"run"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getBunFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getBunFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getBunFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetAngularFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"build"}, false},
		{[]string{"b"}, false},
		{[]string{"test"}, false},
		{[]string{"t"}, false},
		{[]string{"serve"}, false},
		{[]string{"s"}, false},
		{[]string{"lint"}, false},
		{[]string{"generate"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getAngularFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getAngularFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getAngularFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetNxFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{}, false}, // returns filterAutoDetect
		{[]string{"build"}, false},
		{[]string{"run"}, false},
		{[]string{"test"}, false},
		{[]string{"lint"}, false},
		{[]string{"unknown"}, true},
	}
	for _, tc := range cases {
		got := getNxFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getNxFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getNxFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetNpxFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"jest"}, false},
		{[]string{"vitest"}, false},
		{[]string{"mocha"}, false},
		{[]string{"nx", "build"}, false},
		{[]string{"nx"}, false}, // getNxFilter([]) returns filterAutoDetect
		{[]string{"playwright", "test"}, false},
		{[]string{"playwright"}, false},
		{[]string{"playwright", "install"}, true},
		{[]string{"tsc"}, false},
		{[]string{"ng", "build"}, false},
		{[]string{"unknown"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getNpxFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getNpxFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getNpxFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetPipFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"install"}, false},
		{[]string{"list"}, false},
		{[]string{"show"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getPipFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getPipFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getPipFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetUvFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"pip", "install"}, false},
		{[]string{"pip", "list"}, false},
		{[]string{"pip", "show"}, true},
		{[]string{"pip"}, true},
		{[]string{"install"}, false},
		{[]string{"add"}, false},
		{[]string{"unknown"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getUvFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getUvFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getUvFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetBundleFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{}, false}, // returns filterBundleInstall
		{[]string{"install"}, false},
		{[]string{"exec"}, true},
	}
	for _, tc := range cases {
		got := getBundleFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getBundleFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getBundleFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetComposerFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"install"}, false},
		{[]string{"update"}, false},
		{[]string{"require"}, false},
		{[]string{"dump-autoload"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getComposerFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getComposerFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getComposerFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetDotnetFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"build"}, false},
		{[]string{"clean"}, false},
		{[]string{"pack"}, false},
		{[]string{"publish"}, false},
		{[]string{"test"}, false},
		{[]string{"run"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getDotnetFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getDotnetFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getDotnetFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

func TestGetSystemctlFilter(t *testing.T) {
	cases := []struct {
		args    []string
		wantNil bool
	}{
		{[]string{"status"}, false},
		{[]string{"list-units"}, false},
		{[]string{"list-unit-files"}, false},
		{[]string{"start"}, false},
		{[]string{"stop"}, false},
		{[]string{"restart"}, false},
		{[]string{"enable"}, false},
		{[]string{"disable"}, false},
		{[]string{"reload"}, false},
		{[]string{"daemon-reload"}, true},
		{[]string{}, true},
	}
	for _, tc := range cases {
		got := getSystemctlFilter(tc.args)
		if tc.wantNil && got != nil {
			t.Errorf("getSystemctlFilter(%v) = non-nil, want nil", tc.args)
		}
		if !tc.wantNil && got == nil {
			t.Errorf("getSystemctlFilter(%v) = nil, want non-nil", tc.args)
		}
	}
}

// --- skipGitGlobalFlags tests ---

func TestSkipGitGlobalFlags(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "no flags",
			args: []string{"status"},
			want: []string{"status"},
		},
		{
			name: "empty",
			args: []string{},
			want: []string{},
		},
		{
			name: "-C with path",
			args: []string{"-C", "/some/path", "status"},
			want: []string{"status"},
		},
		{
			name: "--git-dir with value",
			args: []string{"--git-dir", "/some/dir", "log"},
			want: []string{"log"},
		},
		{
			name: "--work-tree with value",
			args: []string{"--work-tree", "/some/tree", "diff"},
			want: []string{"diff"},
		},
		{
			name: "-c key=val",
			args: []string{"-c", "user.email=foo@bar.com", "commit"},
			want: []string{"commit"},
		},
		{
			name: "--exec-path with value",
			args: []string{"--exec-path", "/usr/lib/git", "status"},
			want: []string{"status"},
		},
		{
			name: "--git-dir= embedded value",
			args: []string{"--git-dir=/some/dir", "log"},
			want: []string{"log"},
		},
		{
			name: "--work-tree= embedded value",
			args: []string{"--work-tree=/some/tree", "diff"},
			want: []string{"diff"},
		},
		{
			name: "-c= embedded value",
			args: []string{"-c=user.email=foo@bar.com", "status"},
			want: []string{"status"},
		},
		{
			name: "--no-pager",
			args: []string{"--no-pager", "log"},
			want: []string{"log"},
		},
		{
			name: "--bare",
			args: []string{"--bare", "status"},
			want: []string{"status"},
		},
		{
			name: "--no-replace-objects",
			args: []string{"--no-replace-objects", "log"},
			want: []string{"log"},
		},
		{
			name: "-p paginate",
			args: []string{"-p", "log"},
			want: []string{"log"},
		},
		{
			name: "--paginate",
			args: []string{"--paginate", "log"},
			want: []string{"log"},
		},
		{
			name: "multiple flags before subcommand",
			args: []string{"-C", "/path", "--no-pager", "-c", "color.ui=false", "log"},
			want: []string{"log"},
		},
		{
			name: "only flags no subcommand",
			args: []string{"--no-pager", "--bare"},
			want: []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := skipGitGlobalFlags(tc.args)
			if len(got) != len(tc.want) {
				t.Errorf("skipGitGlobalFlags(%v) = %v, want %v", tc.args, got, tc.want)
				return
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("skipGitGlobalFlags(%v)[%d] = %q, want %q", tc.args, i, got[i], tc.want[i])
				}
			}
		})
	}
}

// --- Get() and HasFilter() edge cases ---

func TestGet_EmptyArgs(t *testing.T) {
	// Commands registered without a router return their filter directly.
	// Commands registered with a router that returns nil on empty args should return nil.
	f := Get("git", []string{})
	if f != nil {
		t.Error("Get(git, []) should return nil")
	}
}

func TestGet_UnknownCommand(t *testing.T) {
	f := Get("notacommand", []string{"anything"})
	if f != nil {
		t.Error("Get(notacommand, ...) should return nil")
	}
}

func TestGet_CommandWithDirectFilter(t *testing.T) {
	// ping is a direct filter (no router)
	f := Get("ping", nil)
	if f == nil {
		t.Error("Get(ping, nil) should return non-nil")
	}
}

func TestHasFilter_KnownCommand(t *testing.T) {
	if !HasFilter("git", []string{"status"}) {
		t.Error("HasFilter(git, [status]) should return true")
	}
	if !HasFilter("docker", []string{"ps"}) {
		t.Error("HasFilter(docker, [ps]) should return true")
	}
	if !HasFilter("ping", nil) {
		t.Error("HasFilter(ping, nil) should return true")
	}
}

func TestHasFilter_UnknownCommand(t *testing.T) {
	if HasFilter("notarealcommand", nil) {
		t.Error("HasFilter(notarealcommand, nil) should return false")
	}
}

func TestHasFilter_KnownCommandUnknownSubcommand(t *testing.T) {
	// git with unknown subcommand returns nil from router -> HasFilter false
	if HasFilter("git", []string{"notasubcommand"}) {
		t.Error("HasFilter(git, [notasubcommand]) should return false")
	}
}

// --- safeIdx coverage ---

func TestSafeIdx(t *testing.T) {
	s := []string{"a", "b", "c"}
	if got := safeIdx(s, 0); got != "a" {
		t.Errorf("safeIdx index 0: got %q, want %q", got, "a")
	}
	if got := safeIdx(s, 2); got != "c" {
		t.Errorf("safeIdx index 2: got %q, want %q", got, "c")
	}
	if got := safeIdx(s, -1); got != "" {
		t.Errorf("safeIdx index -1: got %q, want empty", got)
	}
	if got := safeIdx(s, 10); got != "" {
		t.Errorf("safeIdx index 10: got %q, want empty", got)
	}
	if got := safeIdx([]string{}, 0); got != "" {
		t.Errorf("safeIdx empty slice: got %q, want empty", got)
	}
}

// --- filterYarnInstall additional paths ---

func TestFilterYarnInstall_WithErrors(t *testing.T) {
	raw := "yarn install v1.22.19\n" +
		"error An unexpected error occurred.\n" +
		"Done in 1.5s.\n"
	got, err := filterYarnInstall(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "error") {
		t.Errorf("expected error line in output, got: %s", got)
	}
}

func TestFilterYarnInstall_WithWarnings(t *testing.T) {
	raw := "yarn install v1.22.19\n" +
		"warning package-a: No license field\n" +
		"warning package-b: No license field\n" +
		"Done in 2.0s.\n"
	got, err := filterYarnInstall(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "warnings") {
		t.Errorf("expected warnings summary, got: %s", got)
	}
}

func TestFilterYarnInstall_NoDoneLine(t *testing.T) {
	// No done line and no errors/warnings -> result will be empty -> return raw
	raw := "yarn install v1.22.19\n" +
		"[1/4] Resolving packages...\n" +
		"[2/4] Fetching packages...\n"
	got, err := filterYarnInstall(raw)
	if err != nil {
		t.Fatal(err)
	}
	// Should return raw since result is empty
	if got != raw {
		t.Errorf("expected raw passthrough, got: %s", got)
	}
}

func TestFilterYarnInstall_NotYarnOutput(t *testing.T) {
	// looksLikeYarnInstallOutput returns false -> passthrough
	raw := "some other tool output here"
	got, err := filterYarnInstall(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != raw {
		t.Errorf("expected raw passthrough for unrecognized input, got: %s", got)
	}
}

// --- filterUvInstall additional paths ---

func TestFilterUvInstall_WithErrors(t *testing.T) {
	raw := "Resolved 5 packages in 100ms\n" +
		"error: Could not find package foo\n"
	got, err := filterUvInstall(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "error") {
		t.Errorf("expected error line in output, got: %s", got)
	}
}

func TestFilterUvInstall_NoInstallLine(t *testing.T) {
	// Only resolved, no install line -> result empty -> return raw
	raw := "Resolved 5 packages in 100ms\n"
	got, err := filterUvInstall(raw)
	if err != nil {
		t.Fatal(err)
	}
	// No installedLine matched and no errors -> result="" -> return raw
	if got != raw {
		t.Errorf("expected raw passthrough, got: %s", got)
	}
}

func TestFilterUvInstall_NotUvOutput(t *testing.T) {
	raw := "some other tool output here"
	got, err := filterUvInstall(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != raw {
		t.Errorf("expected raw passthrough for unrecognized input, got: %s", got)
	}
}

// --- expandHome additional coverage ---

func TestExpandHome_AdditionalPaths(t *testing.T) {
	// Just "~" alone
	result := expandHome("~")
	if result == "~" {
		// os.UserHomeDir failed, which shouldn't happen in CI - log it
		t.Logf("expandHome(~) returned ~ (UserHomeDir may have failed)")
	} else if result == "" {
		t.Error("expandHome(~) returned empty string")
	}

	// "~/" prefix
	result = expandHome("~/documents")
	if result == "~/documents" {
		t.Logf("expandHome(~/documents) returned unchanged (UserHomeDir may have failed)")
	} else if !strings.Contains(result, "documents") {
		t.Errorf("expandHome(~/documents) = %q, want path containing 'documents'", result)
	}

	// "~\" prefix (Windows style)
	result = expandHome(`~\documents`)
	if result == `~\documents` {
		t.Logf(`expandHome(~\documents) returned unchanged (UserHomeDir may have failed)`)
	} else if !strings.Contains(result, "documents") {
		t.Errorf(`expandHome(~\documents) = %q, want path containing 'documents'`, result)
	}

	// No tilde - should return as-is
	result = expandHome("/absolute/path")
	if result != "/absolute/path" {
		t.Errorf("expandHome(/absolute/path) = %q, want /absolute/path", result)
	}

	// Relative path without tilde
	result = expandHome("relative/path")
	if result != "relative/path" {
		t.Errorf("expandHome(relative/path) = %q, want relative/path", result)
	}
}
