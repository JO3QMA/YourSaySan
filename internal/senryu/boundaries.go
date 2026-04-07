package senryu

import "strings"

// phraseStartOK は 5・7・5 の各句の先頭形態素として自然か（助詞・助動詞で始まらない等）。
func phraseStartOK(m morph) bool {
	switch m.posMajor {
	case "助詞", "助動詞":
		return false
	case "接頭詞":
		return false
	default:
		return true
	}
}

// phraseEndOK は句末として自然か（接頭詞で終わらない、用言は終止・連体・基本形など）。
func phraseEndOK(m morph) bool {
	switch m.posMajor {
	case "接頭詞":
		return false
	case "助詞", "助動詞", "記号", "感動詞":
		return true
	case "名詞", "名詞接続", "形容動詞語幹", "副詞可能", "副詞":
		return true
	case "連体詞":
		return true
	case "接続詞":
		return true
	case "動詞", "形容詞":
		return verbAdjPhraseEndOK(m)
	default:
		return false
	}
}

func verbAdjPhraseEndOK(m morph) bool {
	f := m.inflectionalForm
	if f == "" || f == "*" {
		return false
	}
	// IPA 活用形: 終止形・連体形・基本形・仮定形・命令形・連用タ接続 などを句末として許容
	okForms := []string{
		"終止形", "連体形", "基本形", "仮定形", "命令形",
		"連用タ接続", "連用形", "未然ウ接続",
	}
	for _, o := range okForms {
		if strings.Contains(f, o) {
			return true
		}
	}
	return false
}

// breakBetween は境界の直前形態素 left の直後で切ってよいか（次が right）。
func breakBetween(left, right morph) bool {
	if !phraseEndOK(left) {
		return false
	}
	if !phraseStartOK(right) {
		return false
	}
	// 複合名詞の途中: 名詞 + 名詞（助詞なし）は句切りにしない
	if left.posMajor == "名詞" && right.posMajor == "名詞" {
		return false
	}
	// 名詞 + 接頭詞（次句が接頭詞から始まるのは不自然）
	if right.posMajor == "接頭詞" {
		return false
	}
	return true
}
