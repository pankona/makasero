package detector

import (
	"testing"
)

func TestProposalDetector_IsProposal(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     bool
	}{
		{
			name: "提案を含むレスポンス",
			response: `コードを分析しました。以下の改善を提案します：

---PROPOSAL---
エラーハンドリングを改善します。

---FILE---
internal/api/client.go

---DIFF---
@@ -10,6 +10,7 @@
 func (c *Client) Execute() error {
-    result := process()
+    result, err := process()
+    if err != nil {
+        return fmt.Errorf("実行エラー: %w", err)
+    }
     return nil
 }
---END---`,
			want: true,
		},
		{
			name:     "通常のレスポンス",
			response: "はい、その実装で問題ありません。",
			want:     false,
		},
	}

	detector := NewProposalDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detector.IsProposal(tt.response); got != tt.want {
				t.Errorf("IsProposal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposalDetector_Extract(t *testing.T) {
	tests := []struct {
		name         string
		response     string
		wantErr      bool
		wantNil      bool
		wantProposal *Proposal
	}{
		{
			name: "正常な提案",
			response: `---PROPOSAL---
エラーハンドリングを改善します。
---FILE---
internal/api/client.go
---DIFF---
@@ -10,6 +10,7 @@
 func (c *Client) Execute() error {
-    result := process()
+    result, err := process()
+    if err != nil {
+        return fmt.Errorf("実行エラー: %w", err)
+    }
     return nil
 }
---END---`,
			wantErr: false,
			wantNil: false,
			wantProposal: &Proposal{
				Description: "エラーハンドリングを改善します。",
				FilePath:    "internal/api/client.go",
				Diff: `@@ -10,6 +10,7 @@
 func (c *Client) Execute() error {
-    result := process()
+    result, err := process()
+    if err != nil {
+        return fmt.Errorf("実行エラー: %w", err)
+    }
     return nil
 }`,
			},
		},
		{
			name:         "提案を含まないレスポンス",
			response:     "はい、その実装で問題ありません。",
			wantErr:      false,
			wantNil:      true,
			wantProposal: nil,
		},
		{
			name: "不完全な提案",
			response: `---PROPOSAL---
説明
---FILE---
---DIFF---`,
			wantErr:      true,
			wantNil:      false,
			wantProposal: nil,
		},
	}

	detector := NewProposalDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := detector.Extract(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantNil {
				if got != nil {
					t.Errorf("Extract() = %v, want nil", got)
				}
				return
			}
			if tt.wantProposal != nil {
				if got == nil {
					t.Fatal("Extract() = nil, want non-nil")
				}
				if got.Description != tt.wantProposal.Description {
					t.Errorf("Extract() Description = %v, want %v", got.Description, tt.wantProposal.Description)
				}
				if got.FilePath != tt.wantProposal.FilePath {
					t.Errorf("Extract() FilePath = %v, want %v", got.FilePath, tt.wantProposal.FilePath)
				}
				if got.Diff != tt.wantProposal.Diff {
					t.Errorf("Extract() Diff = %v, want %v", got.Diff, tt.wantProposal.Diff)
				}
			}
		})
	}
}
