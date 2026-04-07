package voicevox

import (
	"encoding/json"
	"strings"
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

func TestFindFirst575Window_nil(t *testing.T) {
	t.Parallel()
	if _, ok := FindFirst575Window(nil); ok {
		t.Fatal("expected false")
	}
}

func TestFindFirst575Window_tooShort(t *testing.T) {
	t.Parallel()
	moras := make([]map[string]any, 16)
	for i := range moras {
		moras[i] = moraJSON("ア", "a")
	}
	q := queryWithMoras(moras, false)
	if _, ok := FindFirst575Window(&q); ok {
		t.Fatal("expected false")
	}
}

func TestFindFirst575Window_embedded18(t *testing.T) {
	t.Parallel()
	moras := make([]map[string]any, 18)
	want := ""
	for i := range moras {
		ch := string(rune('A' + i))
		moras[i] = moraJSON(ch, "a")
		if i < 17 {
			want += ch
		}
	}
	q := queryWithMoras(moras, false)
	got, ok := FindFirst575Window(&q)
	if !ok || got != want {
		t.Fatalf("ok=%v got=%q want=%q", ok, got, want)
	}
}

func TestFindFirst575Window_pauseSplitsSegments(t *testing.T) {
	t.Parallel()
	// 句1: 10 モーラ、pause、句2: 8 モーラ → どちらも 17 未満で部分マッチなし
	first := make([]map[string]any, 10)
	for i := range first {
		first[i] = moraJSON("イ", "i")
	}
	second := make([]map[string]any, 8)
	for i := range second {
		second[i] = moraJSON("ウ", "u")
	}
	wrapped := map[string]any{
		"accent_phrases": []map[string]any{
			{"moras": first, "accent": 1, "isInterrogative": false,
				"pauseMora": map[string]any{"text": "、", "vowel": "pau", "vowel_length": 0.1, "pitch": 0}},
			{"moras": second, "accent": 1, "isInterrogative": false},
		},
		"speedScale": 1, "pitchScale": 0, "intonationScale": 1, "volumeScale": 1,
		"prePhonemeLength": 0.1, "postPhonemeLength": 0.1,
		"outputSamplingRate": 24000, "outputStereo": false,
	}
	raw, _ := json.Marshal(wrapped)
	var q AudioQuery
	if err := json.Unmarshal(raw, &q); err != nil {
		t.Fatal(err)
	}
	if MoraeCountInQuery(&q) != 18 {
		t.Fatalf("total morae got %d", MoraeCountInQuery(&q))
	}
	if _, ok := FindFirst575Window(&q); ok {
		t.Fatal("expected no 17-window within a segment")
	}
	// 全文17（8+9）なら SenryuMatchFromQuery は全体マッチ
	first8 := make([]map[string]any, 8)
	for i := range first8 {
		first8[i] = moraJSON("イ", "i")
	}
	second9 := make([]map[string]any, 9)
	for i := range second9 {
		second9[i] = moraJSON("ウ", "u")
	}
	wrapped2 := map[string]any{
		"accent_phrases": []map[string]any{
			{"moras": first8, "accent": 1, "isInterrogative": false,
				"pauseMora": map[string]any{"text": "、", "vowel": "pau", "vowel_length": 0.1, "pitch": 0}},
			{"moras": second9, "accent": 1, "isInterrogative": false},
		},
		"speedScale": 1, "pitchScale": 0, "intonationScale": 1, "volumeScale": 1,
		"prePhonemeLength": 0.1, "postPhonemeLength": 0.1,
		"outputSamplingRate": 24000, "outputStereo": false,
	}
	raw2, _ := json.Marshal(wrapped2)
	var q2 AudioQuery
	if err := json.Unmarshal(raw2, &q2); err != nil {
		t.Fatal(err)
	}
	if MoraeCountInQuery(&q2) != 17 {
		t.Fatalf("total morae got %d", MoraeCountInQuery(&q2))
	}
	match, ok := SenryuMatchFromQuery(&q2)
	if !ok {
		t.Fatal("expected full 17 match across pause")
	}
	if match != strings.Repeat("イ", 8)+strings.Repeat("ウ", 9) {
		t.Fatalf("match=%q", match)
	}
	if _, ok := FindFirst575Window(&q2); ok {
		t.Fatal("FindFirst575Window should not span segments")
	}
}

func TestSenryuMatchFromQuery_exactly17OnePhrase(t *testing.T) {
	t.Parallel()
	moras := make([]map[string]any, 17)
	for i := range moras {
		moras[i] = moraJSON("x", "a")
	}
	q := queryWithMoras(moras, false)
	match, ok := SenryuMatchFromQuery(&q)
	if !ok || match != strings.Repeat("x", 17) {
		t.Fatalf("ok=%v match=%q", ok, match)
	}
}

func TestJoinAllLinguisticMorae(t *testing.T) {
	t.Parallel()
	q := queryWithMoras([]map[string]any{moraJSON("a", "a"), moraJSON("b", "b")}, true)
	if got := JoinAllLinguisticMorae(&q); got != "ab" {
		t.Fatalf("got %q", got)
	}
}
