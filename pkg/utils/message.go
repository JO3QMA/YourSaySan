package utils

import (
	"regexp"
	"strings"
)

var (
	mentionRegex  = regexp.MustCompile(`<@!?(\d+)>`)
	channelRegex  = regexp.MustCompile(`<#(\d+)>`)
	roleRegex     = regexp.MustCompile(`<@&(\d+)>`)
	emojiRegex    = regexp.MustCompile(`<:(\w+):\d+>`)
	urlRegex      = regexp.MustCompile(`https?://[^\s<>"'()]+`)
	boldRegex     = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicRegex   = regexp.MustCompile(`\*(.+?)\*`)
	underRegex    = regexp.MustCompile(`__(.+?)__`)
	strikeRegex   = regexp.MustCompile(`~~(.+?)~~`)
	inlineCodeRx  = regexp.MustCompile("`(.+?)`")
	codeBlockRx   = regexp.MustCompile("```[\\s\\S]*?```")
	whitespaceRx  = regexp.MustCompile(`\s+`)
)

// ApplyDiscordTextReplacements はメンション・URL・Markdown 等を、読み上げ・川柳判定と同じルールで置換する。
// 改行の空白化・最大長切り詰めは含まない（1行単位の処理では CollapseWhitespace と TrimSpace を併用する）。
// codeBlockRx は引数文字列内でのみマッチする。川柳のように行ごとに呼ぶ場合、複数行にまたがる ``` は除去されない（1行に収まるフェンスのみ効く）。
func ApplyDiscordTextReplacements(content string) string {
	content = mentionRegex.ReplaceAllString(content, "@ユーザー")
	content = channelRegex.ReplaceAllString(content, "#チャンネル")
	content = roleRegex.ReplaceAllString(content, "@ロール")
	content = emojiRegex.ReplaceAllString(content, ":$1:")
	content = urlRegex.ReplaceAllString(content, "URL省略")
	content = boldRegex.ReplaceAllString(content, "$1")
	content = italicRegex.ReplaceAllString(content, "$1")
	content = underRegex.ReplaceAllString(content, "$1")
	content = strikeRegex.ReplaceAllString(content, "$1")
	content = inlineCodeRx.ReplaceAllString(content, "$1")
	content = codeBlockRx.ReplaceAllString(content, "")
	return content
}

// CollapseWhitespace は連続する空白類を1つの半角スペースにまとめる。
func CollapseWhitespace(s string) string {
	return whitespaceRx.ReplaceAllString(s, " ")
}

// TransformMessage はメッセージを読み上げ用に変換する
func TransformMessage(content string, maxLength int) string {
	content = ApplyDiscordTextReplacements(content)

	// 改行を空白に変換
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\r", " ")

	content = CollapseWhitespace(content)
	content = strings.TrimSpace(content)

	// 最大文字数チェック
	if maxLength > 0 && len([]rune(content)) > maxLength {
		runes := []rune(content)
		content = string(runes[:maxLength]) + "以下略"
	}

	return content
}
