package voicevox

import "strings"

// MoraeCountInQuery は VoiceVox の audio_query 応答から言語モーラ数の合計を返す。
// 各アクセント句の Moras スライスのみを数える。PauseMora（vowel が "pau" の休止など）は
// 音韻上のモーラではないため含めない。1語が複数アクセント句に分かれ句間に PauseMora が付く場合も、
// Moras の合計が文学上の音数に一致する（PauseMora を足すと過大になる）。
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
