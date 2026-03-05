package filters

import "testing"

func TestFilterDockerSystemDf(t *testing.T) {
	raw := "TYPE            TOTAL     ACTIVE    SIZE      RECLAIMABLE\n" +
		"Images          15        5         2.5GB     1.8GB (72%)\n" +
		"Containers      8         3         500MB     300MB (60%)\n" +
		"Local Volumes   10        4         1.2GB     800MB (66%)\n" +
		"Build Cache     25        0         3.1GB     3.1GB\n"

	got, err := filterDockerSystemDf(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got == "" {
		t.Error("expected non-empty output")
	}
}

func TestFilterDockerSystemDf_Empty(t *testing.T) {
	got, err := filterDockerSystemDf("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
