package senryu

import "testing"

func TestAnalyzer_ClassicHaiku_threeLines(t *testing.T) {
	t.Parallel()
	a, err := NewAnalyzer()
	if err != nil {
		t.Fatal(err)
	}
	lines := []string{"ふるいけや", "かわずとびこむ", "みずのおと"}
	if !a.CheckThreeLines(lines) {
		t.Fatal("expected classic haiku 3-line match")
	}
}

func TestAnalyzer_ClassicHaiku_blob(t *testing.T) {
	t.Parallel()
	a, err := NewAnalyzer()
	if err != nil {
		t.Fatal(err)
	}
	match, ok := a.FindInBlob("ふるいけやかわずとびこむみずのおと", SenryuBlobMinRunes, 100)
	if !ok {
		t.Fatal("expected blob match")
	}
	if match != "ふるいけやかわずとびこむみずのおと" {
		t.Fatalf("match=%q", match)
	}
}

// 接頭詞を句頭で許可する方針の回帰テスト（経路A: 3行）。
func TestAnalyzer_PrefixAtLineStart_threeLines(t *testing.T) {
	t.Parallel()
	a, err := NewAnalyzer()
	if err != nil {
		t.Fatal(err)
	}
	lines := []string{"お花見の", "桜舞い散る", "春の風"}
	if !a.CheckThreeLines(lines) {
		t.Fatal("expected 5-7-5 with honorific prefix at line start")
	}
}

// 同上（経路B: 1本の blob）。
func TestAnalyzer_PrefixAtLineStart_blob(t *testing.T) {
	t.Parallel()
	a, err := NewAnalyzer()
	if err != nil {
		t.Fatal(err)
	}
	blob := "お花見の桜舞い散る春の風"
	match, ok := a.FindInBlob(blob, SenryuBlobMinRunes, 100)
	if !ok {
		t.Fatal("expected blob match with prefix at segment start")
	}
	if match != blob {
		t.Fatalf("match=%q want %q", match, blob)
	}
}
