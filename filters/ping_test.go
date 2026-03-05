package filters

import (
	"strings"
	"testing"
)

func TestFilterPing(t *testing.T) {
	raw := "PING google.com (142.250.80.46) 56(84) bytes of data.\n" +
		"64 bytes from lax17s62-in-f14.1e100.net (142.250.80.46): icmp_seq=1 ttl=117 time=5.23 ms\n" +
		"64 bytes from lax17s62-in-f14.1e100.net (142.250.80.46): icmp_seq=2 ttl=117 time=4.89 ms\n" +
		"64 bytes from lax17s62-in-f14.1e100.net (142.250.80.46): icmp_seq=3 ttl=117 time=5.12 ms\n\n" +
		"--- google.com ping statistics ---\n" +
		"3 packets transmitted, 3 received, 0% packet loss, time 2004ms\n" +
		"rtt min/avg/max/mdev = 4.890/5.080/5.230/0.127 ms\n"

	got, err := filterPing(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "google.com") {
		t.Error("expected host name")
	}
	if !strings.Contains(got, "0% loss") {
		t.Error("expected loss info")
	}
}

func TestFilterPing_Empty(t *testing.T) {
	got, err := filterPing("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
