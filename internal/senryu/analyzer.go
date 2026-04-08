package senryu

import (
	"fmt"
	"strings"
	"unicode/utf8"

	ipadic "github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

// Analyzer は Kagome IPA 辞書による川柳判定（形態素境界・読みモーラ）。
type Analyzer struct {
	tnz *tokenizer.Tokenizer
}

// NewAnalyzer は IPA 辞書をロードした Analyzer を返す（起動時1回想定）。
func NewAnalyzer() (*Analyzer, error) {
	d := ipadic.Dict()
	tnz, err := tokenizer.New(d, tokenizer.OmitBosEos())
	if err != nil {
		return nil, fmt.Errorf("senryu tokenizer: %w", err)
	}
	return &Analyzer{tnz: tnz}, nil
}

// CheckThreeLines は正規化済み3行が 5・7・5（読みモーラ）かつ句頭・句末が自然かを返す。
func (a *Analyzer) CheckThreeLines(lines []string) bool {
	if len(lines) != 3 {
		return false
	}
	want := []int{5, 7, 5}
	for i, line := range lines {
		if line == "" {
			return false
		}
		ms := a.tokenizeLine(line)
		if len(ms) == 0 {
			return false
		}
		var sum int
		for _, m := range ms {
			sum += m.morae()
		}
		if sum != want[i] {
			return false
		}
		if !phraseStartOK(ms[0]) || !phraseEndOK(ms[len(ms)-1]) {
			return false
		}
	}
	return true
}

// FindInBlob は正規化 blob を Tokenize（Viterbi 1本）し、形態素列上で 5+7+5（17モーラ）かつ品詞境界を満たす最初の連続部分列（表層形連結）を返す。
// Tokenizer は goroutine セーフなため排他不要。
func (a *Analyzer) FindInBlob(blob string, minRunes, maxRunes int) (match string, ok bool) {
	n := utf8.RuneCountInString(blob)
	if n < minRunes || n > maxRunes {
		return "", false
	}

	toks := a.tokenizeLine(blob)
	if len(toks) == 0 {
		return "", false
	}

	targets := []int{5, 7, 5}

	for i := range toks {
		if !phraseStartOK(toks[i]) {
			continue
		}

		var b strings.Builder
		currentPhrase := 0
		moraeSum := 0

		for j := i; j < len(toks); j++ {
			m := toks[j]
			mm := m.morae()
			moraeSum += mm
			b.WriteString(m.surface)

			need := targets[currentPhrase]
			if moraeSum > need {
				break
			}
			if moraeSum < need {
				continue
			}

			// moraeSum == need
			if !phraseEndOK(m) {
				break
			}
			if currentPhrase == 2 {
				return b.String(), true
			}
			if j+1 >= len(toks) {
				break
			}
			if !breakBetween(m, toks[j+1]) {
				break
			}
			currentPhrase++
			moraeSum = 0
		}
	}

	return "", false
}

func (a *Analyzer) tokenizeLine(line string) []morph {
	toks := a.tnz.Tokenize(line)
	var ms []morph
	for _, tok := range toks {
		if tok.Class == tokenizer.DUMMY || tok.Surface == "" {
			continue
		}
		ms = append(ms, morphFromToken(tok))
	}
	return ms
}
