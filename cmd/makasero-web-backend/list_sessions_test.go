package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/pankona/makasero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleListSessions(t *testing.T) {
	// テスト用のセッションディレクトリを作成
	tempDir := t.TempDir()
	originalSessionDir := makasero.SessionDir
	makasero.SessionDir = tempDir
	t.Cleanup(func() {
		makasero.SessionDir = originalSessionDir
	})

	// handleListSessions は sm の内部状態を使わないので、初期化エラーを回避するため空の構造体を使う
	sm := &SessionManager{}
	//sm, err := NewSessionManager() // SessionManager は必要（ハンドラの引数として）
	//require.NoError(t, err, "SessionManager の初期化に成功すべき")

	// --- テストケースの準備 ---
	session1 := &makasero.Session{
		ID:        "session-1",
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-30 * time.Minute),
		// History/SerializedHistory は Marshal/Unmarshal で処理されるので、ここでは空でもOK
	}
	session2 := &makasero.Session{
		ID:        "session-2",
		CreatedAt: time.Now().Add(-2 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	// --- テストシナリオ ---
	tests := []struct {
		name           string
		setupFiles     func(t *testing.T, dir string)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "セッションなし",
			setupFiles: func(t *testing.T, dir string) {
				// 何もしない（ディレクトリは空）
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				assert.JSONEq(t, `[]`, w.Body.String(), "レスポンスは空のJSON配列であるべき")
			},
		},
		{
			name: "複数セッションあり",
			setupFiles: func(t *testing.T, dir string) {
				require.NoError(t, makasero.SaveSession(session1), "session1 の保存に成功すべき")
				require.NoError(t, makasero.SaveSession(session2), "session2 の保存に成功すべき")
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

				var sessions []*makasero.Session
				err := json.Unmarshal(w.Body.Bytes(), &sessions)
				require.NoError(t, err, "レスポンスボディのJSONデコードに成功すべき")
				assert.Len(t, sessions, 2, "2つのセッションが返されるべき")

				// IDでソートして比較しやすくする
				sort.Slice(sessions, func(i, j int) bool {
					return sessions[i].ID < sessions[j].ID
				})

				assert.Equal(t, session1.ID, sessions[0].ID)
				assert.WithinDuration(t, session1.CreatedAt, sessions[0].CreatedAt, time.Second)
				assert.WithinDuration(t, session1.UpdatedAt, sessions[0].UpdatedAt, time.Second)
				// TODO: 必要であれば SerializedHistory の内容も比較

				assert.Equal(t, session2.ID, sessions[1].ID)
				assert.WithinDuration(t, session2.CreatedAt, sessions[1].CreatedAt, time.Second)
				assert.WithinDuration(t, session2.UpdatedAt, sessions[1].UpdatedAt, time.Second)
			},
		},
		{
			name: "不正なJSONファイルとtxtファイルが混在",
			setupFiles: func(t *testing.T, dir string) {
				// 有効なセッション
				require.NoError(t, makasero.SaveSession(session1), "session1 の保存に成功すべき")
				// 不正なJSONファイル
				invalidJsonPath := filepath.Join(dir, "invalid.json")
				require.NoError(t, os.WriteFile(invalidJsonPath, []byte(`{"id": "invalid",`), 0644))
				// JSONではないファイル
				txtPath := filepath.Join(dir, "notes.txt")
				require.NoError(t, os.WriteFile(txtPath, []byte(`this is not json`), 0644))
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

				var sessions []*makasero.Session
				err := json.Unmarshal(w.Body.Bytes(), &sessions)
				require.NoError(t, err, "レスポンスボディのJSONデコードに成功すべき")
				assert.Len(t, sessions, 1, "有効なセッション1つだけが返されるべき")

				assert.Equal(t, session1.ID, sessions[0].ID)
				assert.WithinDuration(t, session1.CreatedAt, sessions[0].CreatedAt, time.Second)
				assert.WithinDuration(t, session1.UpdatedAt, sessions[0].UpdatedAt, time.Second)
			},
		},
		// TODO: ディレクトリ読み込み自体に失敗するケース (パーミッションなど) のテストも考慮できるが、
		//       os.ReadDir のエラーハンドリングが ListSessions 内にあるため、ここでは省略。
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 各テスト前に一時ディレクトリをクリアしてセットアップ
			require.NoError(t, os.RemoveAll(tempDir), "一時ディレクトリのクリアに成功すべき")
			require.NoError(t, os.MkdirAll(tempDir, 0755), "一時ディレクトリの再作成に成功すべき")
			tt.setupFiles(t, tempDir)

			req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
			w := httptest.NewRecorder()

			handleListSessions(w, req, sm)

			assert.Equal(t, tt.expectedStatus, w.Code, "期待されるステータスコードが返されるべき")

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}
