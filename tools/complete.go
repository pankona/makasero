package tools

// CompleteTool はタスクの完了を示すツールです
type CompleteTool struct{}

// Name はツールの名前を返します
func (t *CompleteTool) Name() string {
	return "complete"
}

// Description はツールの説明を返します
func (t *CompleteTool) Description() string {
	return "タスクの完了を示すツールです。実行結果の説明を完了します。"
}

// Execute はツールを実行します
func (t *CompleteTool) Execute(args map[string]interface{}) (string, error) {
	return "タスク完了", nil
}
