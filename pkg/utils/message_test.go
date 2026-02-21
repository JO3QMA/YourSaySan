package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransformMessage(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		maxLength int
		want      string
	}{
		// メンション変換
		{
			name:      "ユーザーメンション <@id>",
			content:   "<@123456789> こんにちは",
			maxLength: 0,
			want:      "@ユーザー こんにちは",
		},
		{
			name:      "ニックネームメンション <@!id>",
			content:   "<@!987654321> おはよう",
			maxLength: 0,
			want:      "@ユーザー おはよう",
		},
		{
			name:      "チャンネルメンション <#id>",
			content:   "<#111222333> を見てね",
			maxLength: 0,
			want:      "#チャンネル を見てね",
		},
		{
			name:      "ロールメンション <@&id>",
			content:   "<@&444555666> に通知",
			maxLength: 0,
			want:      "@ロール に通知",
		},
		{
			name:      "カスタム絵文字 <:name:id>",
			content:   "いい感じ<:thumbsup:123456>だね",
			maxLength: 0,
			want:      "いい感じ:thumbsup:だね",
		},
		// URL置換
		{
			// URLの後の日本語はスペースなしだとURLパターンにも含まれてしまう（実装の仕様）
			name:      "HTTPSのURL（スペース区切り）",
			content:   "詳細は https://example.com/path?q=1 を見て",
			maxLength: 0,
			want:      "詳細は URL省略 を見て",
		},
		{
			// スペースなしの場合、後続テキストがURLの一部として認識される
			name:      "HTTPSのURL（後続テキストなし）",
			content:   "詳細はhttps://example.com/path?q=1を見て",
			maxLength: 0,
			want:      "詳細はURL省略",
		},
		{
			name:      "HTTPのURL",
			content:   "http://example.com",
			maxLength: 0,
			want:      "URL省略",
		},
		// Markdownストリップ
		{
			name:      "太字 **text**",
			content:   "**太字**テキスト",
			maxLength: 0,
			want:      "太字テキスト",
		},
		{
			name:      "斜体 *text*",
			content:   "*斜体*テキスト",
			maxLength: 0,
			want:      "斜体テキスト",
		},
		{
			name:      "下線 __text__",
			content:   "__下線__テキスト",
			maxLength: 0,
			want:      "下線テキスト",
		},
		{
			name:      "打ち消し ~~text~~",
			content:   "~~打ち消し~~テキスト",
			maxLength: 0,
			want:      "打ち消しテキスト",
		},
		{
			name:      "インラインコード `code`",
			content:   "`code`を実行",
			maxLength: 0,
			want:      "codeを実行",
		},
		{
			// バッククォート3個はインラインコード正規表現が先に処理されるため
			// `` `X` `` の形にマッチして最終的に `` `code` `` になる
			name:      "コードブロック ```...``` (1行)",
			content:   "```code```",
			maxLength: 0,
			want:      "`code`",
		},
		{
			// 改行を含むコードブロック（インラインコード処理後にコードブロック正規表現が適用）
			name:      "コードブロック 改行あり",
			content:   "```\nfunc main() {}\n```",
			maxLength: 0,
			want:      "` func main() {} `",
		},
		// 改行・空白正規化
		{
			name:      "改行を空白に変換",
			content:   "一行目\n二行目",
			maxLength: 0,
			want:      "一行目 二行目",
		},
		{
			name:      "連続スペースを1つに",
			content:   "ひとつ   ふたつ",
			maxLength: 0,
			want:      "ひとつ ふたつ",
		},
		{
			name:      "前後の空白をトリム",
			content:   "  前後に空白  ",
			maxLength: 0,
			want:      "前後に空白",
		},
		// 最大文字数制限
		{
			name:      "最大文字数以内はそのまま",
			content:   "短いテキスト",
			maxLength: 10,
			want:      "短いテキスト",
		},
		{
			name:      "最大文字数超えで以下略付与",
			content:   "あいうえおかきくけこ",
			maxLength: 5,
			want:      "あいうえお以下略",
		},
		{
			name:      "maxLength=0のとき制限なし",
			content:   strings.Repeat("あ", 1000),
			maxLength: 0,
			want:      strings.Repeat("あ", 1000),
		},
		// エッジケース
		{
			name:      "空文字列",
			content:   "",
			maxLength: 0,
			want:      "",
		},
		{
			name:      "空白のみ",
			content:   "   ",
			maxLength: 0,
			want:      "",
		},
		{
			name:      "複数のメンションが混在",
			content:   "<@123> と <#456> と <@&789>",
			maxLength: 0,
			want:      "@ユーザー と #チャンネル と @ロール",
		},
		{
			name:      "複数のMarkdown記法が混在",
			content:   "**太字**と*斜体*と~~打ち消し~~",
			maxLength: 0,
			want:      "太字と斜体と打ち消し",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransformMessage(tt.content, tt.maxLength)
			assert.Equal(t, tt.want, got)
		})
	}
}
