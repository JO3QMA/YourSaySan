package senryu

import (
	"github.com/ikawaha/kagome-dict/dict"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	"github.com/ikawaha/kagome/v2/tokenizer/lattice"
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

func morphFromNode(d *dict.Dict, n *lattice.Node) morph {
	var m morph
	m.surface = n.Surface
	switch n.Class {
	case lattice.KNOWN:
		posIDs := d.POSTable.POSs[n.ID]
		if len(posIDs) > 0 {
			m.posMajor = d.POSTable.NameList[posIDs[0]]
		}
		if len(posIDs) > 1 {
			m.posMinor = d.POSTable.NameList[posIDs[1]]
		}
		if len(d.Contents) > n.ID {
			c := d.Contents[n.ID]
			pLen := len(posIDs)
			// tokenizer.FeatureAt と同じ: 素性インデックスから POS 階層長を引いて Contents 列を得る
			if ri := int(ipa.Reading) - pLen; ri >= 0 && ri < len(c) {
				m.reading = c[ri]
			}
			if fi := int(ipa.InflectionalForm) - pLen; fi >= 0 && fi < len(c) {
				m.inflectionalForm = c[fi]
			}
		}
	case lattice.UNKNOWN:
		if len(d.UnkDict.Contents) > n.ID {
			c := d.UnkDict.Contents[n.ID]
			start := 0
			if v, ok := d.UnkDict.ContentsMeta[dict.POSStartIndex]; ok {
				start = int(v)
			}
			end := start + 1
			if v, ok := d.UnkDict.ContentsMeta[dict.POSHierarchy]; ok {
				end = start + int(v)
			}
			if start < len(c) {
				m.posMajor = c[start]
			}
			if start+1 < end && start+1 < len(c) {
				m.posMinor = c[start+1]
			}
		}
	default:
	}
	return m
}
