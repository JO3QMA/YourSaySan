package voicevox

import (
	"encoding/json"
	"testing"
)

func TestMoraeCountInQuery_nil(t *testing.T) {
	t.Parallel()
	if n := MoraeCountInQuery(nil); n != 0 {
		t.Fatalf("got %d", n)
	}
}

func TestMoraeCountInQuery_fromJSON(t *testing.T) {
	t.Parallel()
	raw := `{
		"accent_phrases": [
			{
				"moras": [
					{"text": "ア", "vowel": "a", "vowel_length": 0.1, "pitch": 0},
					{"text": "ア", "vowel": "a", "vowel_length": 0.1, "pitch": 0}
				],
				"accent": 1,
				"isInterrogative": false
			},
			{
				"moras": [
					{"text": "イ", "vowel": "i", "vowel_length": 0.1, "pitch": 0}
				],
				"accent": 1,
				"pauseMora": {"text": "、", "vowel": "pau", "vowel_length": 0.1, "pitch": 0},
				"isInterrogative": false
			}
		],
		"speedScale": 1,
		"pitchScale": 0,
		"intonationScale": 1,
		"volumeScale": 1,
		"prePhonemeLength": 0.1,
		"postPhonemeLength": 0.1,
		"outputSamplingRate": 24000,
		"outputStereo": false
	}`
	var q AudioQuery
	if err := json.Unmarshal([]byte(raw), &q); err != nil {
		t.Fatal(err)
	}
	// pauseMora は言語モーラに含めない: 2 + 1 = 3（PauseMora を数えると 4 になり誤り）
	if n := MoraeCountInQuery(&q); n != 3 {
		t.Fatalf("got %d, want 3", n)
	}
}

func TestMoraeCountInQuery_ignoresPauseMoraBetweenPhrases(t *testing.T) {
	t.Parallel()
	// 1語が2アクセント句に分かれ、2句目にのみ pauseMora がある想定。Moras 合計 5 のまま。
	raw := `{
		"accent_phrases": [
			{
				"moras": [
					{"text": "フ", "vowel": "u", "vowel_length": 0.1, "pitch": 0},
					{"text": "ル", "vowel": "u", "vowel_length": 0.1, "pitch": 0},
					{"text": "イ", "vowel": "i", "vowel_length": 0.1, "pitch": 0}
				],
				"accent": 1,
				"isInterrogative": false
			},
			{
				"moras": [
					{"text": "ケ", "vowel": "e", "vowel_length": 0.1, "pitch": 0},
					{"text": "ヤ", "vowel": "a", "vowel_length": 0.1, "pitch": 0}
				],
				"accent": 1,
				"pauseMora": {"text": "、", "vowel": "pau", "vowel_length": 0.1, "pitch": 0},
				"isInterrogative": false
			}
		],
		"speedScale": 1,
		"pitchScale": 0,
		"intonationScale": 1,
		"volumeScale": 1,
		"prePhonemeLength": 0.1,
		"postPhonemeLength": 0.1,
		"outputSamplingRate": 24000,
		"outputStereo": false
	}`
	var q AudioQuery
	if err := json.Unmarshal([]byte(raw), &q); err != nil {
		t.Fatal(err)
	}
	if n := MoraeCountInQuery(&q); n != 5 {
		t.Fatalf("got %d, want 5 (must not add PauseMora as a 6th mora)", n)
	}
}

func moraJSON(text, vowel string) map[string]any {
	return map[string]any{
		"text": text, "vowel": vowel, "vowel_length": 0.1, "pitch": 0,
	}
}

func queryWithMoras(moras []map[string]any, pauseAfterPhrase bool) AudioQuery {
	phrases := []map[string]any{
		{
			"moras":           moras,
			"accent":          1,
			"isInterrogative": false,
		},
	}
	if pauseAfterPhrase && len(moras) > 0 {
		phrases[0]["pauseMora"] = map[string]any{"text": "、", "vowel": "pau", "vowel_length": 0.1, "pitch": 0}
	}
	wrapped := map[string]any{
		"accent_phrases":     phrases,
		"speedScale":         1,
		"pitchScale":         0,
		"intonationScale":    1,
		"volumeScale":        1,
		"prePhonemeLength":   0.1,
		"postPhonemeLength":  0.1,
		"outputSamplingRate": 24000,
		"outputStereo":       false,
	}
	raw, err := json.Marshal(wrapped)
	if err != nil {
		panic(err)
	}
	var q AudioQuery
	if err := json.Unmarshal(raw, &q); err != nil {
		panic(err)
	}
	return q
}

func TestJoinAllLinguisticMorae(t *testing.T) {
	t.Parallel()
	q := queryWithMoras([]map[string]any{moraJSON("a", "a"), moraJSON("b", "b")}, true)
	if got := JoinAllLinguisticMorae(&q); got != "ab" {
		t.Fatalf("got %q", got)
	}
}
