package filters

import (
	"strings"
	"testing"
)

func TestFilterCmake_Configure(t *testing.T) {
	raw := "-- The C compiler identification is GNU 12.3.0\n" +
		"-- Detecting C compiler ABI info\n" +
		"-- Detecting C compiler ABI info - done\n" +
		"-- Configuring done (1.2s)\n" +
		"-- Generating done (0.1s)\n" +
		"-- Build files have been written to: /home/user/project/build\n"

	got, err := filterCmake(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "configured in 1.2s") {
		t.Errorf("expected config time, got: %s", got)
	}
}

func TestFilterCmake_Build(t *testing.T) {
	raw := "[ 25%] Building CXX object CMakeFiles/myapp.dir/src/main.cpp.o\n" +
		"[ 50%] Building CXX object CMakeFiles/myapp.dir/src/utils.cpp.o\n" +
		"[ 75%] Building CXX object CMakeFiles/myapp.dir/src/config.cpp.o\n" +
		"[100%] Linking CXX executable myapp\n" +
		"[100%] Built target myapp\n"

	got, err := filterCmake(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "built target myapp") {
		t.Errorf("expected target, got: %s", got)
	}
}

func TestFilterCmake_Empty(t *testing.T) {
	got, err := filterCmake("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
