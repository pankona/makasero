# コマンド推論機能の調査結果

## 概要

Roo-Codeのコマンド実行機能の調査結果と、それに基づく我々の実装方針をまとめます。

## Roo-Codeの実装方式

### 1. 基本アーキテクチャ

Roo-Codeは以下の方式でコマンド実行を制御しています：

- 特別なコマンド推論機能は実装せず、LLMの基本的な能力に依存
- システムプロンプトと動作規則による制御
- ツールグループとモードによる機能制限
- 実行前のシステム情報確認の義務付け

### 2. コンポーネント構成

#### 2.1 モード制御（`modes.ts`）
```typescript
export const modes: readonly ModeConfig[] = [
  {
    slug: "code",
    name: "Code",
    roleDefinition: "You are Roo, a highly skilled software engineer...",
    groups: ["read", "edit", "browser", "command", "mcp"],
  },
  // ... 他のモード
]
```

#### 2.2 ツールグループ（`tool-groups.ts`）
```typescript
export const TOOL_GROUPS: Record<string, ToolGroupValues> = {
  read: ["read_file", "search_files", "list_files", ...],
  edit: ["write_to_file", "apply_diff"],
  command: ["execute_command"],
  // ... 他のグループ
}
```

#### 2.3 実行ツール（`execute-command.ts`）
```typescript
export function getExecuteCommandDescription(args: ToolArgs): string {
  return `## execute_command
Description: Request to execute a CLI command on the system...
Parameters:
- command: (required) The CLI command to execute...`
}
```

### 3. 制御メカニズム

#### 3.1 システムプロンプトによる制御
- 複数のセクションから構成
- 各セクションが特定の機能や制約を定義
- `rules.ts`で詳細な動作規則を定義

#### 3.2 実行前の確認事項
1. システム情報の確認
2. 作業ディレクトリの制約
3. コマンドの安全性チェック
4. 実行結果の解釈方法

#### 3.3 安全性の考慮
- 現在の作業ディレクトリでの実行に限定
- 有害な指示の禁止
- OSに適したコマンドの確認

### 4. 実装の詳細

#### 4.1 コマンド実行フロー
1. ユーザー入力の受付
2. LLMによるコマンド生成
3. 安全性チェック
4. コマンド実行
5. 結果の解釈と返却

#### 4.2 エラーハンドリング
1. 実行前のエラー
   - コマンドの構文エラー
   - 権限エラー
   - 安全性チェックエラー
2. 実行時のエラー
   - プロセス実行エラー
   - タイムアウト
   - リソース制限
3. 実行後のエラー
   - 出力解析エラー
   - 結果の検証エラー

#### 4.3 セキュリティ対策
1. コマンドの検証
   - 危険なコマンド（rm, mv等）のブロック
   - パスの検証（作業ディレクトリ外へのアクセス防止）
   - 引数のサニタイズ
2. 実行環境の制限
   - 実行ユーザーの権限制限
   - リソース制限（CPU, メモリ, 実行時間）
   - ネットワークアクセスの制限

#### 4.4 結果の解釈
1. 出力形式
   ```typescript
   interface CommandResult {
     success: boolean;
     output: string;
     error?: string;
     exitCode: number;
     duration: number;
   }
   ```
2. LLMへの結果提供
   ```
   コマンド実行結果：
   - 成功: [true/false]
   - 出力: [標準出力]
   - エラー: [標準エラー出力]
   - 終了コード: [数値]
   ```
3. ユーザーへの表示
   - 成功/失敗の明確な表示
   - エラーの場合の原因説明
   - 必要に応じた出力の整形

## 我々の実装への示唆

### 1. 採用すべき点
1. システムプロンプトによる基本的な制御
2. モードとツールグループによる機能制限
3. 詳細な実行規則の定義

### 2. 改善・拡張の余地
1. より構造化されたコマンド推論機能の追加
2. 安全性チェックの強化
3. 実行結果の解釈機能の強化

### 3. 実装の優先順位
1. 基本的なプロンプト構造の実装
2. コマンド実行の安全性チェック
3. 実行結果の解釈機能
4. 高度なコマンド推論機能

### 4. 実装上の注意点
1. エラーハンドリング
   - 各段階でのエラー処理の実装
   - ユーザーフレンドリーなエラーメッセージ
   - リカバリー手順の提供
2. セキュリティ
   - 包括的な安全性チェック
   - 実行環境の適切な制限
   - 監査ログの実装
3. パフォーマンス
   - コマンド実行の非同期処理
   - タイムアウト制御
   - リソース使用の最適化

## 参考資料
- Roo-Code実装（`src/core/prompts/`）
- システムプロンプト定義（`src/core/prompts/system.ts`）
- ツールグループ定義（`src/shared/tool-groups.ts`）
- 実行規則（`src/core/prompts/sections/rules.ts`） 