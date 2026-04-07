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
