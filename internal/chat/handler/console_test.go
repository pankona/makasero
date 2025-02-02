package handler

import (
	"os"
	"testing"

	"github.com/pankona/makasero/internal/chat/detector"
	"github.com/stretchr/testify/assert"
)

func TestNewConsoleApprover(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "正常系：ConsoleApproverの作成",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			approver := NewConsoleApprover()
			assert.NotNil(t, approver)
			assert.NotNil(t, approver.tty)
		})
	}
}

func TestConsoleApprover_GetApproval(t *testing.T) {
	// テスト用のパイプを作成
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer r.Close()
	defer w.Close()

	approver := &ConsoleApprover{
		tty: r,
	}

	proposal := &detector.Proposal{
		Description: "テスト提案",
		FilePath:    "test.go",
		Diff:        "テストコード",
	}

	tests := []struct {
		name    string
		input   string
		want    bool
		wantErr bool
	}{
		{
			name:    "正常系：承認",
			input:   "y\n",
			want:    true,
			wantErr: false,
		},
		{
			name:    "正常系：大文字での承認",
			input:   "Y\n",
			want:    true,
			wantErr: false,
		},
		{
			name:    "正常系：拒否",
			input:   "n\n",
			want:    false,
			wantErr: false,
		},
		{
			name:    "正常系：空入力での拒否",
			input:   "\n",
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用の入力を書き込む
			go func() {
				w.Write([]byte(tt.input))
			}()

			got, err := approver.GetApproval(proposal)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
