package senryu

import (
	"context"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/JO3QMA/YourSaySan/pkg/utils"
)

// 改行なし経路B・文中検出: 短すぎる blob を除外し VoiceVox 呼び出しを抑制する下限ルーン数
const SenryuBlobMinRunes = 12

// 改行なし経路B: 17モーラ相当の短文に絞り VoiceVox 呼び出しを抑制（VOICEVOX_MAX_MESSAGE_LENGTH=50 に近い上限）
const unbrokenSenryuMaxRunes = 40

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

// NormalizeSenryuBlob は Discord 向け置換のあと、Unicode 空白をすべて除去した1本の文字列を返す。
func NormalizeSenryuBlob(content string) string {
	s := utils.ApplyDiscordTextReplacements(content)
	var b strings.Builder
	for _, r := range s {
		if !unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// IsUnbrokenSenryuCandidate は本文に改行がなく、正規化後の blob が川柳（連続17モーラ）候補として妥当な長さのとき (blob, true)。
func IsUnbrokenSenryuCandidate(content string) (blob string, ok bool) {
	if strings.ContainsAny(content, "\n\r") {
		return "", false
	}
	blob = NormalizeSenryuBlob(content)
	if blob == "" {
		return "", false
	}
	n := utf8.RuneCountInString(blob)
	if n < SenryuBlobMinRunes || n > unbrokenSenryuMaxRunes {
		return "", false
	}
	return blob, true
}

// Is575MoraeUnbroken は連続1文の合計モーラ数がちょうど17かどうかを返す（改行なし経路B）。
func Is575MoraeUnbroken(ctx context.Context, c MoraeCounter, blob string, speakerID int) (bool, error) {
	if blob == "" {
		return false, nil
	}
	n, err := c.CountMorae(ctx, blob, speakerID)
	if err != nil {
		return false, err
	}
	return n == 5+7+5, nil
}

// FormatSenryuReply は SENRYU_REPLY_TEXT にマッチ箇所を埋め込む。テンプレートに %s がちょうど1つあるときは fmt.Sprintf、それ以外は末尾に「…」を付与する。
func FormatSenryuReply(template, match string) string {
	if strings.Count(template, "%s") == 1 {
		return fmt.Sprintf(template, match)
	}
	return template + "\n「" + match + "」"
}
