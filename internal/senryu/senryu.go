package senryu

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/JO3QMA/YourSaySan/pkg/utils"
)

// 改行なし経路B・文中検出: 短すぎる blob を除外し、Kagome 解析の無駄な走査を減らす下限ルーン数
const SenryuBlobMinRunes = 12

// 改行なし経路B: 17モーラ相当の短文に絞り、解析コストと誤検知リスクを抑える（VOICEVOX_MAX_MESSAGE_LENGTH=50 に近い上限）
const unbrokenSenryuMaxRunes = 40

// NormalizeLine は1行を川柳判定・読み上げ前処理と同種の置換を行う（改行は含まない想定）。
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

func isSentenceDelimiterRune(r rune) bool {
	switch r {
	case '。', '\uFF0E', '?', '？', '!', '！':
		return true
	default:
		return false
	}
}

// SplitBlobBySentenceDelimiters は正規化済み blob を句点・疑問符・感嘆符で分割する（区切り文字は結果に含めない）。
// 区切りが無いときは []string{blob} を返す。空入力・区切りのみの入力は nil。
func SplitBlobBySentenceDelimiters(blob string) []string {
	if blob == "" {
		return nil
	}
	var out []string
	var seg strings.Builder
	for _, r := range blob {
		if isSentenceDelimiterRune(r) {
			if seg.Len() > 0 {
				out = append(out, seg.String())
				seg.Reset()
			}
			continue
		}
		seg.WriteRune(r)
	}
	if seg.Len() > 0 {
		out = append(out, seg.String())
	}
	if len(out) == 0 {
		return nil
	}
	return out
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

// FormatSenryuReply は SENRYU_REPLY_TEXT にマッチ箇所を埋め込む。テンプレートに %s がちょうど1つあるときは fmt.Sprintf、それ以外は末尾に「…」を付与する。
func FormatSenryuReply(template, match string) string {
	if strings.Count(template, "%s") == 1 {
		return fmt.Sprintf(template, match)
	}
	return template + "\n「" + match + "」"
}
