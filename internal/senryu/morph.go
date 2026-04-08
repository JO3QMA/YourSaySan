package senryu

import (
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

// morph は品詞境界判定とモーラ集計用の最小情報。
type morph struct {
	surface          string
	reading          string
	posMajor         string
	posMinor         string
	inflectionalForm string
}

func (m morph) morae() int {
	if m.reading != "" && m.reading != "*" {
		n := CountMoraeInReading(m.reading)
		if n > 0 {
			return n
		}
	}
	// UNKNOWN 等で読みが無いときは表面から推定するが、漢字1字1モーラ等で実読みとずれることがある（検出品質の限界）。
	return moraeFromSurface(m.surface)
}

func morphFromToken(tok tokenizer.Token) morph {
	var m morph
	m.surface = tok.Surface
	pos := tok.POS()
	if len(pos) > 0 {
		m.posMajor = pos[0]
	}
	if len(pos) > 1 {
		m.posMinor = pos[1]
	}
	if r, ok := tok.FeatureAt(int(ipa.Reading)); ok && r != "" && r != "*" {
		m.reading = r
	}
	if f, ok := tok.FeatureAt(int(ipa.InflectionalForm)); ok {
		m.inflectionalForm = f
	}
	return m
}
