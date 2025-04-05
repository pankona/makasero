package tools

// Tool はエージェントが使用できるツールのインターフェースを定義します
type Tool interface {
	// Name はツールの名前を返します
	Name() string

	// Description はツールの説明を返します
	Description() string

	// Execute はツールを実行し、結果を返します
	Execute(args map[string]interface{}) (string, error)
}
