package handler

import (
	"errors"
	"testing"

	"github.com/pankona/makasero/internal/chat/detector"
)

// モックApprover
type mockApprover struct {
	approved bool
	err      error
}

func (m *mockApprover) GetApproval(*detector.Proposal) (bool, error) {
	return m.approved, m.err
}

// モックApplier
type mockApplier struct {
	err error
}

func (m *mockApplier) Apply(*detector.Proposal) error {
	return m.err
}

func TestProposalHandler_Handle(t *testing.T) {
	testProposal := &detector.Proposal{
		Description: "テスト提案",
		FilePath:    "test/file.go",
		Diff:        "テスト差分",
	}

	tests := []struct {
		name     string
		approver *mockApprover
		applier  *mockApplier
		wantErr  bool
	}{
		{
			name: "正常系：承認あり",
			approver: &mockApprover{
				approved: true,
				err:      nil,
			},
			applier: &mockApplier{
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "正常系：承認なし",
			approver: &mockApprover{
				approved: false,
				err:      nil,
			},
			applier: &mockApplier{
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "異常系：承認エラー",
			approver: &mockApprover{
				approved: false,
				err:      errors.New("承認エラー"),
			},
			applier: &mockApplier{
				err: nil,
			},
			wantErr: true,
		},
		{
			name: "異常系：適用エラー",
			approver: &mockApprover{
				approved: true,
				err:      nil,
			},
			applier: &mockApplier{
				err: errors.New("適用エラー"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewProposalHandler(tt.approver, tt.applier)
			err := h.Handle(testProposal)

			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
