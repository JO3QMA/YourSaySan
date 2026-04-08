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
	// 無区切りのひらがな長連続は Viterbi 上で1語化されやすく 5+7+5 の形態素境界と一致しない。
	// 読点で区切ると経路Bでも検出できる（Discord 正規化は空白のみ除去し、読点は残る）。
	blob := "ふるいけや、かわずとびこむ、みずのおと"
	match, ok := a.FindInBlob(blob, SenryuBlobMinRunes, 100)
	if !ok {
		t.Fatal("expected blob match")
	}
	if match != blob {
		t.Fatalf("match=%q want %q", match, blob)
	}
}

// 無区切り古典俳句の連続ひらがなは経路Bでは検出しない（トークン境界の制約）。
func TestAnalyzer_ClassicHaiku_blob_noDelimiter_notMatched(t *testing.T) {
	t.Parallel()
	a, err := NewAnalyzer()
	if err != nil {
		t.Fatal(err)
	}
	blob := "ふるいけやかわずとびこむみずのおと"
	if _, ok := a.FindInBlob(blob, SenryuBlobMinRunes, 100); ok {
		t.Fatal("expected no blob match for undelimited kana run")
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
	// FindInBlob は Tokenize 1本の形態素列上で最初に見つかった 5+7+5（表層形連結）を返す。
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

func TestAnalyzer_FindInBlob_noFalsePositive_morphChat(t *testing.T) {
	t.Parallel()
	a, err := NewAnalyzer()
	if err != nil {
		t.Fatal(err)
	}
	blob := "一応形態素解析するようにはしたけど"
	if _, ok := a.FindInBlob(blob, SenryuBlobMinRunes, 100); ok {
		t.Fatal("expected no match for morph-analysis chat line (former DFS false positive)")
	}
}

func TestAnalyzer_FindInBlob_noFalsePositive_gengo(t *testing.T) {
	t.Parallel()
	a, err := NewAnalyzer()
	if err != nil {
		t.Fatal(err)
	}
	blob := "言語がなんで分かれるんだろう"
	if _, ok := a.FindInBlob(blob, SenryuBlobMinRunes, 100); ok {
		t.Fatal("expected no match (former DFS split 言/語 false positive)")
	}
}
