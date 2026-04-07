package senryu

import (
	"context"
	"strings"

	"github.com/JO3QMA/YourSaySan/pkg/utils"
)

// MoraeCounter は VoiceVox 等による1行あたりのモーラ数取得に使う。
type MoraeCounter interface {
	CountMorae(ctx context.Context, text string, speakerID int) (int, error)
}

// NormalizeLine は1行を VoiceVox に渡す前に、TransformMessage と同種の置換を行う（改行は含まない想定）。
// コードフェンスが複数行に分かれる場合は行単位では utils の codeBlock 除去が効かない（意図せぬ過剰除去も避けられる）。
func NormalizeLine(s string) string {
	s = utils.ApplyDiscordTextReplacements(s)
	s = utils.CollapseWhitespace(s)
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
