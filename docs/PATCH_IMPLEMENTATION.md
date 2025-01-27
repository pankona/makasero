# パッチ適用システムの実装設計

## 1. パッチ適用の基本設計

### パッチ情報の構造
```go
// internal/patch/types.go
type PatchInfo struct {
    FilePath    string   // 対象ファイルパス
    HunkRanges []Range  // 変更範囲情報
    Lines      []Line   // 変更行情報
}

type Range struct {
    StartLine int
    LineCount int
}

type Line struct {
    Type     LineType // 追加/削除/コンテキスト
    Content  string   // 行の内容
    Number   int      // 行番号
}

type LineType int

const (
    LineContext LineType = iota
    LineAdd
    LineRemove
)
```

### パッチパーサー
```go
// internal/patch/parser.go
type PatchParser struct {
    scanner *bufio.Scanner
}

func (p *PatchParser) Parse(diff string) (*PatchInfo, error) {
    // 1. @@ -start,count +start,count @@ 形式のヘッダー解析
    // 2. 変更行の解析（+, -, スペース）
    // 3. PatchInfo構造体の生成
}
```

## 2. パッチ適用プロセス

### ファイル操作ユーティリティ
```go
// internal/patch/fileutil.go
type FileUtil struct {
    backupDir string
}

// バックアップの作成
func (f *FileUtil) CreateBackup(filePath string) (string, error) {
    backupPath := filepath.Join(f.backupDir, 
        fmt.Sprintf("%s.%d.bak", 
            filepath.Base(filePath), 
            time.Now().UnixNano()))
    
    return backupPath, copyFile(filePath, backupPath)
}

// ファイルの復元
func (f *FileUtil) RestoreFromBackup(backupPath, originalPath string) error {
    return copyFile(backupPath, originalPath)
}
```

### パッチ適用器
```go
// internal/patch/applier.go
type PatchApplier struct {
    fileUtil *FileUtil
}

func (a *PatchApplier) Apply(patch *PatchInfo) error {
    // 1. ファイルの存在確認
    if !fileExists(patch.FilePath) {
        return ErrFileNotFound
    }

    // 2. バックアップの作成
    backupPath, err := a.fileUtil.CreateBackup(patch.FilePath)
    if err != nil {
        return fmt.Errorf("バックアップ作成エラー: %w", err)
    }

    // 3. 一時ファイルの作成
    tempFile, err := createTempFile()
    if err != nil {
        return fmt.Errorf("一時ファイル作成エラー: %w", err)
    }
    defer os.Remove(tempFile.Name())

    // 4. パッチの適用
    if err := a.applyPatch(patch, tempFile); err != nil {
        // エラー時はバックアップから復元
        a.fileUtil.RestoreFromBackup(backupPath, patch.FilePath)
        return fmt.Errorf("パッチ適用エラー: %w", err)
    }

    // 5. 一時ファイルを元のファイルに移動
    if err := os.Rename(tempFile.Name(), patch.FilePath); err != nil {
        a.fileUtil.RestoreFromBackup(backupPath, patch.FilePath)
        return fmt.Errorf("ファイル置換エラー: %w", err)
    }

    return nil
}

func (a *PatchApplier) applyPatch(patch *PatchInfo, output *os.File) error {
    input, err := os.Open(patch.FilePath)
    if err != nil {
        return err
    }
    defer input.Close()

    scanner := bufio.NewScanner(input)
    currentLine := 1
    
    for _, hunk := range patch.HunkRanges {
        // 1. ハンク開始位置までコピー
        for currentLine < hunk.StartLine {
            if !scanner.Scan() {
                return ErrUnexpectedEOF
            }
            fmt.Fprintln(output, scanner.Text())
            currentLine++
        }

        // 2. ハンク内の変更を適用
        for _, line := range patch.Lines {
            switch line.Type {
            case LineAdd:
                fmt.Fprintln(output, line.Content)
            case LineRemove:
                if !scanner.Scan() {
                    return ErrUnexpectedEOF
                }
                if scanner.Text() != line.Content {
                    return ErrPatchMismatch
                }
                currentLine++
            case LineContext:
                if !scanner.Scan() {
                    return ErrUnexpectedEOF
                }
                fmt.Fprintln(output, scanner.Text())
                currentLine++
            }
        }
    }

    // 3. 残りのファイル内容をコピー
    for scanner.Scan() {
        fmt.Fprintln(output, scanner.Text())
    }

    return scanner.Err()
}
```

## 3. エラー処理

### カスタムエラー
```go
// internal/patch/errors.go
var (
    ErrFileNotFound   = errors.New("対象ファイルが見つかりません")
    ErrPatchMismatch  = errors.New("パッチがファイルの内容と一致しません")
    ErrUnexpectedEOF  = errors.New("予期せぬファイル終端")
    ErrInvalidHunk    = errors.New("無効なハンク範囲")
)

type PatchError struct {
    Phase   string // パース、適用、バックアップなど
    Message string
    Err     error
}

func (e *PatchError) Error() string {
    return fmt.Sprintf("%s: %s (%v)", e.Phase, e.Message, e.Err)
}
```

## 4. テスト戦略

### ユニットテスト
```go
// internal/patch/applier_test.go
func TestPatchApplier(t *testing.T) {
    cases := []struct {
        name     string
        original string
        patch    *PatchInfo
        want     string
        wantErr  error
    }{
        {
            name: "単純な行追加",
            original: "line1\nline2\n",
            patch: &PatchInfo{
                HunkRanges: []Range{{StartLine: 2, LineCount: 1}},
                Lines: []Line{
                    {Type: LineAdd, Content: "new line"},
                },
            },
            want: "line1\nline2\nnew line\n",
        },
        // その他のテストケース
    }
    // テスト実装
}
```

### 統合テスト
```go
// internal/patch/integration_test.go
func TestPatchWorkflow(t *testing.T) {
    // 1. テストファイルの作成
    // 2. パッチの作成と適用
    // 3. 結果の検証
    // 4. エラーケースの検証
}
```

## 5. 安全性の考慮事項

1. ファイルシステムの安全性
- バックアップの作成
- アトミックな操作
- 権限の確認

2. エラー時の回復
- バックアップからの復元
- 一時ファイルの削除
- トランザクション的な操作

3. 並行処理の考慮
- ファイルロック
- 競合の検出
- 順序の保証

## 6. パフォーマンス最適化

1. メモリ使用量
- ストリーム処理
- バッファサイズの最適化
- 大きなファイルの処理

2. 処理速度
- 効率的なアルゴリズム
- キャッシュの活用
- 不要なI/O操作の削減

## 7. 拡張性

1. 新しいパッチフォーマット
```go
// internal/patch/format.go
type PatchFormat interface {
    Parse(input string) (*PatchInfo, error)
    Generate(original, modified string) (string, error)
}
```

2. カスタムバックアップ戦略
```go
// internal/patch/backup.go
type BackupStrategy interface {
    CreateBackup(path string) (string, error)
    RestoreBackup(backupPath, originalPath string) error
    CleanupBackup(backupPath string) error
}
```

3. プログレス通知
```go
// internal/patch/progress.go
type ProgressReporter interface {
    OnStart(total int)
    OnProgress(current int)
    OnComplete()
    OnError(err error)
}