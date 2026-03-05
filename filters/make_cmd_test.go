package filters

import (
	"strings"
	"testing"
)

func TestFilterMake(t *testing.T) {
	raw := "make[1]: Entering directory '/home/user/project/src'\n" +
		"gcc -c -Wall -O2 -o main.o main.c\n" +
		"gcc -c -Wall -O2 -o utils.o utils.c\n" +
		"main.c:45:10: warning: unused variable 'temp' [-Wunused-variable]\n" +
		"gcc -o myapp main.o utils.o -lm\n" +
		"make[1]: Leaving directory '/home/user/project/src'\n"

	got, err := filterMake(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "build ok") {
		t.Error("expected build ok")
	}
	if !strings.Contains(got, "warning") {
		t.Error("expected warning preserved")
	}
}

func TestFilterMake_Empty(t *testing.T) {
	got, err := filterMake("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
