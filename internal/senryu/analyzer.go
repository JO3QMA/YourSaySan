package senryu

import (
	"fmt"
	"sync"
	"unicode/utf8"

	"github.com/ikawaha/kagome-dict/dict"
	ipadic "github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

// Analyzer は Kagome IPA 辞書による川柳判定（形態素境界・読みモーラ）。
type Analyzer struct {
	mu  sync.Mutex
	tnz *tokenizer.Tokenizer
	dic *dict.Dict
}

// NewAnalyzer は IPA 辞書をロードした Analyzer を返す（起動時1回想定）。
func NewAnalyzer() (*Analyzer, error) {
	d := ipadic.Dict()
	tnz, err := tokenizer.New(d, tokenizer.OmitBosEos())
	if err != nil {
		return nil, fmt.Errorf("senryu tokenizer: %w", err)
	}
	return &Analyzer{tnz: tnz, dic: d}, nil
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

// FindInBlob は正規化 blob 内で 5+7+5（17モーラ）かつ品詞境界を満たす最初の部分文字列を返す。
// 辞書は読み取り専用のため a.mu は不要（Tokenizer は共有しない）。
func (a *Analyzer) FindInBlob(blob string, minRunes, maxRunes int) (match string, ok bool) {
	n := utf8.RuneCountInString(blob)
	if n < minRunes || n > maxRunes {
		return "", false
	}

	// blob 経路は読み取り専用の辞書参照のみ（arcsFromPosition）。Tokenizer と mutex は共有しない。
	targets := []int{5, 7, 5}
	rc := n

	for s := 0; s < rc; s++ {
		if end, found := dfsBlobMatch(a.dic, blob, rc, s, targets); found {
			return substringByRunes(blob, s, end), true
		}
	}
	return "", false
}

func (a *Analyzer) tokenizeLine(line string) []morph {
	a.mu.Lock()
	defer a.mu.Unlock()
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

type dfsMemoKey struct {
	pos, segIdx, acc int
	lastPOS, lastInfl string
}

// dfsBlobMatch は複数分割候補を列挙する（arcsFromPosition は lattice.Build と同じ弧集合）。
func dfsBlobMatch(d *dict.Dict, inp string, rc, start int, targets []int) (endRune int, ok bool) {
	memo := make(map[dfsMemoKey]struct {
		end int
		ok  bool
	})

	var dfs func(pos, segIdx, acc int, lastSegEnd *morph) (int, bool)
	dfs = func(pos, segIdx, acc int, lastSegEnd *morph) (int, bool) {
		if segIdx >= len(targets) {
			return pos, true
		}
		lPOS, lInfl := "", ""
		if lastSegEnd != nil {
			lPOS = lastSegEnd.posMajor
			lInfl = lastSegEnd.inflectionalForm
		}
		mk := dfsMemoKey{pos: pos, segIdx: segIdx, acc: acc, lastPOS: lPOS, lastInfl: lInfl}
		if v, ok := memo[mk]; ok {
			return v.end, v.ok
		}
		save := func(end int, ok bool) (int, bool) {
			memo[mk] = struct {
				end int
				ok  bool
			}{end, ok}
			return end, ok
		}

		need := targets[segIdx]

		var resultEnd int
		var found bool

		for _, arc := range arcsFromPosition(d, nil, inp, pos) {
			end, node := arc.end, arc.node
			if end > rc {
				continue
			}
			m := morphFromNode(d, node)
			mm := m.morae()
			if mm == 0 || acc+mm > need {
				continue
			}

			if acc == 0 {
				if !phraseStartOK(m) {
					continue
				}
				if segIdx > 0 && lastSegEnd != nil && !breakBetween(*lastSegEnd, m) {
					continue
				}
			}

			if acc+mm == need {
				if !phraseEndOK(m) {
					continue
				}
				mc := m
				if segIdx == len(targets)-1 {
					resultEnd = end
					found = true
					break
				}
				e, subOK := dfs(end, segIdx+1, 0, &mc)
				if subOK {
					resultEnd = e
					found = true
					break
				}
				continue
			}

			e, subOK := dfs(end, segIdx, acc+mm, lastSegEnd)
			if subOK {
				resultEnd = e
				found = true
				break
			}
		}

		if found {
			return save(resultEnd, true)
		}
		return save(0, false)
	}

	return dfs(start, 0, 0, nil)
}
