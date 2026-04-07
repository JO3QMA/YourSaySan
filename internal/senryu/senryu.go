package senryu

import (
	"context"
	"regexp"
	"strings"
)

// MoraeCounter は VoiceVox 等による1行あたりのモーラ数取得に使う。
type MoraeCounter interface {
	CountMorae(ctx context.Context, text string, speakerID int) (int, error)
}

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

// NormalizeLine は1行を VoiceVox に渡す前に、TransformMessage と同種の置換を行う（改行は含まない想定）。
func NormalizeLine(s string) string {
	s = mentionRegex.ReplaceAllString(s, "@ユーザー")
	s = channelRegex.ReplaceAllString(s, "#チャンネル")
	s = roleRegex.ReplaceAllString(s, "@ロール")
	s = emojiRegex.ReplaceAllString(s, ":$1:")
	s = urlRegex.ReplaceAllString(s, "URL省略")
	s = boldRegex.ReplaceAllString(s, "$1")
	s = italicRegex.ReplaceAllString(s, "$1")
	s = underRegex.ReplaceAllString(s, "$1")
	s = strikeRegex.ReplaceAllString(s, "$1")
	s = inlineCodeRx.ReplaceAllString(s, "$1")
	s = codeBlockRx.ReplaceAllString(s, "")
	s = whitespaceRx.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// ThreeLines は空行を除いた非空行がちょうど3行のときだけ正規化済みの行を返す。
func ThreeLines(content string) (lines []string, ok bool) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	parts := strings.Split(content, "\n")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, NormalizeLine(p))
	}
	if len(out) != 3 {
		return nil, false
	}
	for _, line := range out {
		if line == "" {
			return nil, false
		}
	}
	return out, true
}

// Is575Morae は3行がそれぞれモーラ数 5, 7, 5 かどうかを返す。
func Is575Morae(ctx context.Context, c MoraeCounter, lines []string, speakerID int) (bool, error) {
	if len(lines) != 3 {
		return false, nil
	}
	want := []int{5, 7, 5}
	for i, line := range lines {
		n, err := c.CountMorae(ctx, line, speakerID)
		if err != nil {
			return false, err
		}
		if n != want[i] {
			return false, nil
		}
	}
	return true, nil
}
