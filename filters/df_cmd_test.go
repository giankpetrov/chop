package filters

import (
	"strings"
	"testing"
)

func TestFilterDf(t *testing.T) {
	raw := "Filesystem      Size  Used Avail Use% Mounted on\n" +
		"udev            7.8G     0  7.8G   0% /dev\n" +
		"tmpfs           1.6G  2.1M  1.6G   1% /run\n" +
		"/dev/sda1       100G   45G   50G  48% /\n" +
		"tmpfs           7.8G     0  7.8G   0% /dev/shm\n" +
		"/dev/sdb1       500G  200G  275G  42% /data\n"

	got, err := filterDf(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if strings.Contains(got, "tmpfs") {
		t.Error("expected tmpfs filtered out")
	}
	if !strings.Contains(got, "/dev/sda1") {
		t.Error("expected real filesystem kept")
	}
}

func TestFilterDf_Empty(t *testing.T) {
	got, err := filterDf("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
