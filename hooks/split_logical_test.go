package hooks

import (
	"reflect"
	"testing"
)

func TestSplitLogical(t *testing.T) {
	tests := []struct {
		name          string
		command       string
		wantSegments  []string
		wantOperators []string
	}{
		{
			name:          "simple &&",
			command:       "git add . && git commit",
			wantSegments:  []string{"git add .", "git commit"},
			wantOperators: []string{" && "},
		},
		{
			name:          "simple ||",
			command:       "npm install || echo failed",
			wantSegments:  []string{"npm install", "echo failed"},
			wantOperators: []string{" || "},
		},
		{
			name:          "simple ;",
			command:       "cd /tmp ; ls",
			wantSegments:  []string{"cd /tmp", "ls"},
			wantOperators: []string{" ; "},
		},
		{
			name:          "multiple operators",
			command:       "cmd1 && cmd2 || cmd3 ; cmd4",
			wantSegments:  []string{"cmd1", "cmd2", "cmd3", "cmd4"},
			wantOperators: []string{" && ", " || ", " ; "},
		},
		{
			name:          "&& inside double quotes",
			command:       `grep "foo && bar" file.txt`,
			wantSegments:  []string{`grep "foo && bar" file.txt`},
			wantOperators: nil,
		},
		{
			name:          "|| inside single quotes",
			command:       `grep 'feat || fix' logs.txt`,
			wantSegments:  []string{`grep 'feat || fix' logs.txt`},
			wantOperators: nil,
		},
		{
			name:          "; inside quotes",
			command:       `echo "hello ; world"`,
			wantSegments:  []string{`echo "hello ; world"`},
			wantOperators: nil,
		},
		{
			name:          "escaped double quote inside double quotes",
			command:       `grep "say \" && write" file.txt && echo done`,
			wantSegments:  []string{`grep "say \" && write" file.txt`, "echo done"},
			wantOperators: []string{" && "},
		},
		{
			name:          "mixed quoted and unquoted",
			command:       `grep "a && b" file && cmd2 || echo "c || d"`,
			wantSegments:  []string{`grep "a && b" file`, "cmd2", `echo "c || d"`},
			wantOperators: []string{" && ", " || "},
		},
		{
			name:          "no operators",
			command:       "ls -la",
			wantSegments:  []string{"ls -la"},
			wantOperators: nil,
		},
		{
			name:          "empty string",
			command:       "",
			wantSegments:  []string{""},
			wantOperators: nil,
		},
		{
			name:          "leading operator (technically invalid shell but should split)",
			command:       " && cmd",
			wantSegments:  []string{"", "cmd"},
			wantOperators: []string{" && "},
		},
		{
			name:          "trailing operator",
			command:       "cmd && ",
			wantSegments:  []string{"cmd", ""},
			wantOperators: []string{" && "},
		},
		{
			name:          "operators without spaces (should not split according to current implementation)",
			command:       "cmd1&&cmd2",
			wantSegments:  []string{"cmd1&&cmd2"},
			wantOperators: nil,
		},
		{
			name:          "multiple spaces around operators",
			command:       "cmd1  &&  cmd2",
			wantSegments:  []string{"cmd1 ", " cmd2"}, // Note: currently it splits on exactly " && "
			wantOperators: []string{" && "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSegments, gotOperators := splitLogical(tt.command)
			if !reflect.DeepEqual(gotSegments, tt.wantSegments) {
				t.Errorf("splitLogical() gotSegments = %v, want %v", gotSegments, tt.wantSegments)
			}
			if !reflect.DeepEqual(gotOperators, tt.wantOperators) {
				t.Errorf("splitLogical() gotOperators = %v, want %v", gotOperators, tt.wantOperators)
			}
		})
	}
}
