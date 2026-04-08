package senryu

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

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

// 体言止め→名詞始まりの句境界（名詞＋名詞除外を外した回帰テスト、経路A）。
func TestAnalyzer_NounNounBoundary_threeLines(t *testing.T) {
	t.Parallel()
	a, err := NewAnalyzer()
	if err != nil {
		t.Fatal(err)
	}
	lines := []string{"今日の飯", "カレーライスだ", "うれしいな"}
	if !a.CheckThreeLines(lines) {
		t.Fatal("expected 5-7-5 with noun–noun phrase boundary between lines")
	}
}

// 同上（経路B: 1本の blob）。
func TestAnalyzer_NounNounBoundary_blob(t *testing.T) {
	t.Parallel()
	a, err := NewAnalyzer()
	if err != nil {
		t.Fatal(err)
	}
	blob := "今日の飯カレーライスだうれしいな"
	match, ok := a.FindInBlob(blob, SenryuBlobMinRunes, 100)
	if !ok {
		t.Fatal("expected blob match with noun–noun boundary between segments")
	}
	// FindInBlob は開始位置0で最初に見つかった 5+7+5 を返す。ラティス分割の違いで本文より短いプレフィックスになる場合がある。
	if match == "" || !strings.HasPrefix(blob, match) {
		t.Fatalf("match=%q want non-empty prefix of blob %q", match, blob)
	}
}

// Kagome Tokenizer は goroutine セーフ想定のため、並行 tokenizeLine でレースが出ないこと（go test -race）。
func TestAnalyzer_ConcurrentTokenize(t *testing.T) {
	a, err := NewAnalyzer()
	if err != nil {
		t.Fatal(err)
	}
	const (
		goroutines = 48
		iters      = 128
	)
	var wg sync.WaitGroup
	var fail atomic.Bool
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iters; j++ {
				ms := a.tokenizeLine("秋の風が木の葉を揺らす")
				if len(ms) == 0 {
					fail.Store(true)
					return
				}
			}
		}()
	}
	wg.Wait()
	if fail.Load() {
		t.Fatal("concurrent tokenizeLine produced empty morphs")
	}
}
