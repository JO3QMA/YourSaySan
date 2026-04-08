package senryu

import (
	"unicode/utf8"

	"github.com/ikawaha/kagome-dict/dict"
	"github.com/ikawaha/kagome/v2/tokenizer/lattice"
)

// kagome の lattice.Build と同じ分岐・順序で、runeStart から出る弧を列挙する。
// 上流の *lattice.Lattice に ArcsFrom が無いため、辞書走査で同等の候補を得る。
const maximumUnknownWordLength = 1024

type latticeArc struct {
	end  int
	node *lattice.Node
}

func byteOffsetAtRuneIndex(inp string, runeIdx int) (byteOff int, ok bool) {
	if runeIdx < 0 {
		return 0, false
	}
	r := 0
	for pos := range inp {
		if r == runeIdx {
			return pos, true
		}
		r++
	}
	return 0, false
}

func makeLatticeNode(d *dict.Dict, runePos, bytePos, id int, class lattice.NodeClass, surface string) *lattice.Node {
	n := &lattice.Node{
		ID:       id,
		Position: bytePos,
		Start:    runePos,
		Class:    class,
		Surface:  surface,
	}
	var m dict.Morph
	switch class {
	case lattice.KNOWN:
		m = d.Morphs[id]
	case lattice.UNKNOWN:
		m = d.UnkDict.Morphs[id]
	case lattice.USER:
		// lattice.addNode と同様（デフォルトコスト）
	}
	n.Left, n.Right, n.Weight = int32(m.LeftID), int32(m.RightID), int32(m.Weight)
	return n
}

func arcsFromPosition(d *dict.Dict, u *dict.UserDict, inp string, runeStart int) []latticeArc {
	pos, ok := byteOffsetAtRuneIndex(inp, runeStart)
	if !ok {
		return nil
	}
	ch, _ := utf8.DecodeRuneInString(inp[pos:])
	var out []latticeArc
	anyMatches := false

	if u != nil {
		u.Index.CommonPrefixSearchCallback(inp[pos:], func(id, l int) {
			surf := inp[pos : pos+l]
			end := runeStart + utf8.RuneCountInString(surf)
			out = append(out, latticeArc{
				end:  end,
				node: makeLatticeNode(d, runeStart, pos, id, lattice.USER, surf),
			})
			anyMatches = true
		})
	}
	if anyMatches {
		return out
	}

	d.Index.CommonPrefixSearchCallback(inp[pos:], func(id, l int) {
		surf := inp[pos : pos+l]
		end := runeStart + utf8.RuneCountInString(surf)
		out = append(out, latticeArc{
			end:  end,
			node: makeLatticeNode(d, runeStart, pos, id, lattice.KNOWN, surf),
		})
		anyMatches = true
	})

	class := d.CharacterCategory(ch)
	if !anyMatches || d.InvokeList[int(class)] {
		var endPos int
		if ch != utf8.RuneError {
			endPos = pos + utf8.RuneLen(ch)
		} else {
			endPos = pos + 1
		}
		unkWordLen := 1
		if d.GroupList[int(class)] {
			for i, w, size := endPos, 0, len(inp); i < size; i += w {
				var c rune
				c, w = utf8.DecodeRuneInString(inp[i:])
				if d.CharacterCategory(c) != class {
					break
				}
				endPos += w
				unkWordLen++
				if unkWordLen >= maximumUnknownWordLength {
					break
				}
			}
		}

		prev := pos
		if c, size := utf8.DecodeLastRuneInString(inp[pos:endPos]); c != utf8.RuneError {
			prev = endPos - size
		}
		id := d.UnkDict.Index[int32(class)]
		dup := d.UnkDict.IndexDup[int32(class)]
		for x := range int(dup) + 1 {
			if pos < prev {
				surf := inp[pos:prev]
				end := runeStart + utf8.RuneCountInString(surf)
				out = append(out, latticeArc{
					end:  end,
					node: makeLatticeNode(d, runeStart, pos, int(id)+x, lattice.UNKNOWN, surf),
				})
			}
			surf := inp[pos:endPos]
			end := runeStart + utf8.RuneCountInString(surf)
			out = append(out, latticeArc{
				end:  end,
				node: makeLatticeNode(d, runeStart, pos, int(id)+x, lattice.UNKNOWN, surf),
			})
		}
	}

	return out
}
