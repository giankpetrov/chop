package updater

import (
	"fmt"
	"strings"
	"testing"
)

func BenchmarkParseChecksum(b *testing.B) {
	// Generate a large checksums string
	numLines := 1000
	var sb strings.Builder
	for i := 0; i < numLines; i++ {
		fmt.Fprintf(&sb, "hash%d  chop-linux-amd64-%d\n", i, i)
	}
	// Add the one we're looking for at the end
	binaryName := "chop-target-binary"
	fmt.Fprintf(&sb, "targethash  %s\n", binaryName)
	checksums := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash, err := parseChecksum(checksums, binaryName)
		if err != nil {
			b.Fatal(err)
		}
		if hash != "targethash" {
			b.Fatalf("expected targethash, got %s", hash)
		}
	}
}
