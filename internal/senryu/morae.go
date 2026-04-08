package senryu

import (
	"strings"
	"unicode"
)

// CountMoraeInReading はカタカナ（またはひらがな）の読みからモーラ数を数える。
// 拗音（ゃゅょぁぃぅぇぉ）は直前の音と合わせて1拍。促音っ・撥音ん・長音ーは各1拍。
func CountMoraeInReading(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	s = toKatakanaReading(s)
	var runes = []rune(s)
	n := 0
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if isSmallKana(r) {
			if n == 0 {
				n++
			}
			continue
		}
		if isKana(r) || r == 'ン' || r == 'ッ' || r == 'ー' {
			n++
			continue
		}
		if unicode.IsLetter(r) {
			n++
		}
	}
	return n
}

func toKatakanaReading(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'ぁ' && r <= 'ゖ':
			b.WriteRune(r + 0x60)
		case r >= 'ァ' && r <= 'ヺ':
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isSmallKana(r rune) bool {
	switch r {
	case 'ァ', 'ィ', 'ゥ', 'ェ', 'ォ', 'ャ', 'ュ', 'ョ', 'ヮ',
		'ぁ', 'ぃ', 'ぅ', 'ぇ', 'ぉ', 'ゃ', 'ゅ', 'ょ', 'ゎ':
		return true
	default:
		return false
	}
}

func isKana(r rune) bool {
	return (r >= 'ァ' && r <= 'ヺ') || (r >= 'ぁ' && r <= 'ゖ')
}

// moraeFromSurface は読みが空のとき表層形をひらがな相当としてモーラ数の近似に使う。
func moraeFromSurface(surface string) int {
	return CountMoraeInReading(surface)
}
