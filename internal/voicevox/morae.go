package voicevox

// MoraeCountInQuery は VoiceVox の audio_query 応答からモーラ数の合計を返す。
// 文学上の厳密な音数ではなく、エンジンの分解に基づく数（読み上げと整合する）。
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
