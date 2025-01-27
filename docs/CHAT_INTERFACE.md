# チャットインターフェース設計

## 1. インターフェース設計

### 1.1 コマンドライン引数
```bash
roo -command chat -input '[{"role":"user","content":"Hello"}]'
roo -command explain -input "fmt.Println('Hello, World!')"
```

### 1.2 出力フォーマット
```json
{
  "success": true,
  "data": "応答内容",
  "error": null
}
```

## 2. コマンド仕様

### 2.1 chatコマンド
```bash
# 単一メッセージ
roo -command chat -input '[
  {"role":"user","content":"こんにちは"}
]'

# 会話の文脈を含むメッセージ
roo -command chat -input '[
  {"role":"system","content":"あなたはプログラミング講師です"},
  {"role":"user","content":"Goの並行処理について教えてください"},
  {"role":"assistant","content":"Goの並行処理は..."},
  {"role":"user","content":"もう少し詳しく説明してください"}
]'
```

### 2.2 explainコマンド
```bash
# 直接コードを指定
roo -command explain -input "func add(a, b int) int { return a + b }"

# ファイルからコードを読み込んで説明
cat main.go | roo -command explain -input "$(cat)"
```

## 3. 実装詳細

### 3.1 メッセージ処理
```go
// メッセージ構造
type ChatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

// チャットリクエスト
type ChatRequest struct {
    Model    string        `json:"model"`
    Messages []ChatMessage `json:"messages"`
}

// レスポンス構造
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}
```

### 3.2 APIクライアント
```go
// APIクライアント
type Client struct {
    httpClient *http.Client
    apiKey     string
    baseURL    string
}

// チャット完了リクエスト
func (c *Client) CreateChatCompletion(messages []ChatMessage) (string, error) {
    req := &ChatRequest{
        Model:    "gpt-4",
        Messages: messages,
    }
    
    // APIリクエストの送信と応答の処理
    return c.sendRequest(req)
}
```

## 4. エラーハンドリング

### 4.1 入力検証
```go
func validateInput(input string) error {
    if input == "" {
        return errors.New("input is required")
    }
    return nil
}
```

### 4.2 APIエラー処理
```go
func handleAPIError(resp *http.Response) error {
    var errResp ErrorResponse
    if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
        return fmt.Errorf("API error (status %d)", resp.StatusCode)
    }
    return fmt.Errorf("API error: %s", errResp.Error.Message)
}
```

## 5. 使用例

### 5.1 シェルスクリプトとの連携
```bash
#!/bin/bash

# コードレビュー支援
review_code() {
    git diff | roo -command explain -input "$(cat)"
}

# チャットセッション
chat_session() {
    message=$1
    history_file=".chat_history.json"
    
    if [ ! -f "$history_file" ]; then
        echo "[]" > "$history_file"
    fi
    
    # 履歴を読み込んで新しいメッセージを追加
    new_history=$(jq --arg msg "$message" '. + [{"role":"user","content":$msg}]' "$history_file")
    response=$(echo "$new_history" | roo -command chat -input "$(cat)")
    
    # 応答を履歴に追加
    echo "$response" | jq -r '.data' | \
        jq --arg content "$(cat)" '. + [{"role":"assistant","content":$content}]' > "$history_file"
}
```

### 5.2 パイプラインでの使用
```bash
# ソースコードの説明
find . -name "*.go" -exec cat {} \; | roo -command explain -input "$(cat)"

# コミットメッセージの生成
git diff | roo -command chat -input '[
  {"role":"system","content":"コミットメッセージを生成してください"},
  {"role":"user","content":"'"$(cat)"'"}
]' | jq -r .data
```

## 6. パフォーマンス最適化

### 6.1 タイムアウト設定
```go
const defaultTimeout = 30 * time.Second

client := &http.Client{
    Timeout: defaultTimeout,
}
```

### 6.2 メモリ効率
```go
// 大きな入力の効率的な処理
func processLargeInput(reader io.Reader) error {
    scanner := bufio.NewScanner(reader)
    scanner.Buffer(make([]byte, 1024*1024), 1024*1024*10)
    
    for scanner.Scan() {
        // 1行ずつ処理
    }
    return scanner.Err()
}
```

## 7. テスト方針

### 7.1 ユニットテスト
```go
func TestCreateChatCompletion(t *testing.T) {
    client := NewTestClient(t)
    messages := []ChatMessage{
        {Role: "user", Content: "Hello"},
    }
    
    response, err := client.CreateChatCompletion(messages)
    assert.NoError(t, err)
    assert.NotEmpty(t, response)
}
```

### 7.2 統合テスト
```go
func TestEndToEnd(t *testing.T) {
    cmd := exec.Command("roo",
        "-command", "chat",
        "-input", `[{"role":"user","content":"Hello"}]`)
    
    output, err := cmd.CombinedOutput()
    assert.NoError(t, err)
    
    var response Response
    assert.NoError(t, json.Unmarshal(output, &response))
    assert.True(t, response.Success)
}