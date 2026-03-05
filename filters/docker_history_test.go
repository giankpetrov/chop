package filters

import (
	"strings"
	"testing"
)

func TestFilterDockerHistory(t *testing.T) {
	raw := "IMAGE          CREATED        CREATED BY                                      SIZE      COMMENT\n" +
		"abc123def456   2 days ago     /bin/sh -c #(nop)  CMD [\"node\" \"server.js\"]      0B\n" +
		"<missing>      2 days ago     /bin/sh -c npm install --production               150MB\n" +
		"<missing>      2 days ago     /bin/sh -c #(nop) COPY dir:abc123 in /app         5.2MB\n" +
		"<missing>      2 weeks ago    /bin/sh -c #(nop)  CMD [\"node\"]                   0B\n" +
		"<missing>      2 weeks ago    /bin/sh -c apt-get update && apt-get install       250MB\n"

	got, err := filterDockerHistory(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if strings.Contains(got, "0B") {
		t.Error("expected 0B layers to be skipped")
	}
	if !strings.Contains(got, "layers") {
		t.Error("expected layer count")
	}
}

func TestFilterDockerHistory_Empty(t *testing.T) {
	got, err := filterDockerHistory("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
