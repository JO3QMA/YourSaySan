package utils

import (
	"regexp"
	"strings"
)

// TransformMessage はメッセージを読み上げ用に変換する
func TransformMessage(content string, maxLength int) string {
	// 1. メンション変換（<@user_id> -> @username）
	// 実際の実装では、Discord APIからユーザー名を取得する必要がある
	// ここでは簡易実装として、メンションを「ユーザー」に変換
	mentionRegex := regexp.MustCompile(`<@!?(\d+)>`)
	content = mentionRegex.ReplaceAllString(content, "@ユーザー")

	// 2. チャンネルメンション変換（<#channel_id> -> #channel-name）
	channelRegex := regexp.MustCompile(`<#(\d+)>`)
	content = channelRegex.ReplaceAllString(content, "#チャンネル")

	// 3. ロールメンション変換（<@&role_id> -> @role-name）
	roleRegex := regexp.MustCompile(`<@&(\d+)>`)
	content = roleRegex.ReplaceAllString(content, "@ロール")

	// 4. カスタム絵文字変換（<:name:id> -> :name:）
	emojiRegex := regexp.MustCompile(`<:(\w+):\d+>`)
	content = emojiRegex.ReplaceAllString(content, ":$1:")

	// 5. URLを「URL省略」に置換
	urlRegex := regexp.MustCompile(`https?://[^\s<>"'()]+`)
	content = urlRegex.ReplaceAllString(content, "URL省略")

	// 6. Markdown記法の除去
	// **bold** -> bold
	content = regexp.MustCompile(`\*\*(.+?)\*\*`).ReplaceAllString(content, "$1")
	// *italic* -> italic
	content = regexp.MustCompile(`\*(.+?)\*`).ReplaceAllString(content, "$1")
	// __underline__ -> underline
	content = regexp.MustCompile(`__(.+?)__`).ReplaceAllString(content, "$1")
	// ~~strikethrough~~ -> strikethrough
	content = regexp.MustCompile(`~~(.+?)~~`).ReplaceAllString(content, "$1")
	// `code` -> code
	content = regexp.MustCompile("`(.+?)`").ReplaceAllString(content, "$1")
	// ```code block``` -> code block
	content = regexp.MustCompile("```[\\s\\S]*?```").ReplaceAllString(content, "")

	// 7. 改行を空白に変換
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\r", " ")

	// 8. 連続する空白を1つに
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")

	// 9. 前後の空白を削除
	content = strings.TrimSpace(content)

	// 10. 最大文字数チェック
	if maxLength > 0 && len([]rune(content)) > maxLength {
		runes := []rune(content)
		content = string(runes[:maxLength]) + "以下略"
	}

	return content
}

