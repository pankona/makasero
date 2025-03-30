package tools

// Complete はタスク完了を示します
type Complete struct {
	Summary string
}

// Execute はタスク完了のサマリーを返します
func (c *Complete) Execute() (string, error) {
	return c.Summary, nil
} 