package voicevox

import "strings"

// MoraeCountInQuery は VoiceVox の audio_query 応答から言語モーラ数の合計を返す。
// 各アクセント句の Moras スライスのみを数える。PauseMora（vowel が "pau" の休止など）は
// 音韻上のモーラではないため含めない。1語が複数アクセント句に分かれ句間に PauseMora が付く場合も、
// Moras の合計が文学上の音数に一致する（PauseMora を足すと過大になる）。
// 文学上の厳密な音数ではなく、エンジンの分解に基づく数（読み上げ・川柳判定と整合する）。
func MoraeCountInQuery(q *AudioQuery) int {
	if q == nil {
		return 0
	}
	n := 0
	for i := range q.AccentPhrases {
		ap := &q.AccentPhrases[i]
		n += len(ap.Moras)
	}
	return n
}

// accentPhraseSegments は PauseMora があるアクセント句の直後でセグメントを切る。
// 句読点をまたいだ連続17モーラの誤検知を減らすため、セグメント内の連続モーラのみを対象にする。
func accentPhraseSegments(q *AudioQuery) [][]Mora {
	if q == nil {
		return nil
	}
	var segments [][]Mora
	var cur []Mora
	for i := range q.AccentPhrases {
		ap := &q.AccentPhrases[i]
		cur = append(cur, ap.Moras...)
		if ap.PauseMora != nil {
			segments = append(segments, cur)
			cur = nil
		}
	}
	if len(cur) > 0 {
		segments = append(segments, cur)
	}
	return segments
}

func joinMoraTexts(moras []Mora) string {
	var b strings.Builder
	for i := range moras {
		b.WriteString(moras[i].Text)
	}
	return b.String()
}

// JoinAllLinguisticMorae は全アクセント句の Moras を順に連結した表示用文字列を返す（PauseMora は含めない）。
func JoinAllLinguisticMorae(q *AudioQuery) string {
	if q == nil {
		return ""
	}
	var b strings.Builder
	for i := range q.AccentPhrases {
		ap := &q.AccentPhrases[i]
		for j := range ap.Moras {
			b.WriteString(ap.Moras[j].Text)
		}
	}
	return b.String()
}

// FindFirst575Window は PauseMora 区切りセグメントごとに、連続17モーラ（5+7+5）の先頭ウィンドウを探す。
func FindFirst575Window(q *AudioQuery) (match string, ok bool) {
	if q == nil {
		return "", false
	}
	for _, seg := range accentPhraseSegments(q) {
		if len(seg) < 17 {
			continue
		}
		return joinMoraTexts(seg[:17]), true
	}
	return "", false
}

// SenryuMatchFromQuery は全文がちょうど17モーラなら全体を、そうでなければセグメント内の先頭17モーラを返す。
func SenryuMatchFromQuery(q *AudioQuery) (match string, ok bool) {
	if q == nil {
		return "", false
	}
	if MoraeCountInQuery(q) == 17 {
		return JoinAllLinguisticMorae(q), true
	}
	return FindFirst575Window(q)
}
