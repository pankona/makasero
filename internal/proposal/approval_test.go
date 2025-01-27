package proposal

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockStdin は標準入力をモックするためのヘルパー関数
func mockStdin(input string) func() {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	io.WriteString(w, input)
	w.Close()

	return func() {
		os.Stdin = oldStdin
	}
}

func TestCLIApprover_RequestApproval(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		proposal  *CodeProposal
		want      bool
		wantError bool
	}{
		{
			name:  "正常系：承認",
			input: "y\n",
			proposal: &CodeProposal{
				OriginalCode: "Hello World",
				ProposedCode: "Hello Go",
				FilePath:     "test.txt",
				DiffContent:  "@@ -1 +1 @@\n-Hello World\n+Hello Go\n",
				ApplyMode:    ApplyModePatch,
				Description:  "World を Go に変更",
			},
			want:      true,
			wantError: false,
		},
		{
			name:  "正常系：拒否",
			input: "n\n",
			proposal: &CodeProposal{
				OriginalCode: "Hello World",
				ProposedCode: "Hello Go",
				FilePath:     "test.txt",
				DiffContent:  "@@ -1 +1 @@\n-Hello World\n+Hello Go\n",
				ApplyMode:    ApplyModePatch,
				Description:  "World を Go に変更",
			},
			want:      false,
			wantError: false,
		},
		{
			name:  "正常系：大文字での承認",
			input: "Y\n",
			proposal: &CodeProposal{
				OriginalCode: "Hello World",
				ProposedCode: "Hello Go",
				FilePath:     "test.txt",
				DiffContent:  "@@ -1 +1 @@\n-Hello World\n+Hello Go\n",
				ApplyMode:    ApplyModePatch,
				Description:  "World を Go に変更",
			},
			want:      true,
			wantError: false,
		},
		{
			name:  "正常系：大文字での拒否",
			input: "N\n",
			proposal: &CodeProposal{
				OriginalCode: "Hello World",
				ProposedCode: "Hello Go",
				FilePath:     "test.txt",
				DiffContent:  "@@ -1 +1 @@\n-Hello World\n+Hello Go\n",
				ApplyMode:    ApplyModePatch,
				Description:  "World を Go に変更",
			},
			want:      false,
			wantError: false,
		},
		{
			name:  "正常系：yesでの承認",
			input: "yes\n",
			proposal: &CodeProposal{
				OriginalCode: "Hello World",
				ProposedCode: "Hello Go",
				FilePath:     "test.txt",
				DiffContent:  "@@ -1 +1 @@\n-Hello World\n+Hello Go\n",
				ApplyMode:    ApplyModePatch,
				Description:  "World を Go に変更",
			},
			want:      true,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 標準入力のモック
			cleanup := mockStdin(tt.input)
			defer cleanup()

			approver := NewCLIApprover()
			got, err := approver.RequestApproval(tt.proposal)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
