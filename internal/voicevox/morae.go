package voicevox

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
