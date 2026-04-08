package senryu

import (
	"strings"
	"testing"
)

func TestThreeLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantOK  bool
		wantLen int
	}{
		{
			name:    "three lines no blanks",
			input:   "いちごはなか\nにじいろのゆめ\nみどりのかぜ",
			wantOK:  true,
			wantLen: 3,
		},
		{
			name: "three lines with blank lines ignored",
			input: `あああああ


いいいいいいい
ううううう`,
			wantOK:  true,
			wantLen: 3,
		},
		{
			name:    "crlf",
			input:   "line1\r\nline2\r\nline3",
			wantOK:  true,
			wantLen: 3,
		},
		{
			name:    "four lines",
			input:   "a\nb\nc\nd",
			wantOK:  false,
			wantLen: 0,
		},
		{
			name:    "one line",
			input:   "only one",
			wantOK:  false,
			wantLen: 0,
		},
		{
			name:    "two lines",
			input:   "a\nb",
			wantOK:  false,
			wantLen: 0,
		},
		{
			name:    "empty after normalize",
			input:   "   \n\t\n   ",
			wantOK:  false,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := ThreeLines(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if tt.wantOK && len(got) != tt.wantLen {
				t.Fatalf("len(lines) = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestNormalizeLine_Mention(t *testing.T) {
	t.Parallel()
	got := NormalizeLine("hello <@123456> world")
	want := "hello @ユーザー world"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNormalizeSenryuBlob_stripsSpaceAndNewline(t *testing.T) {
	t.Parallel()
	// 本文に改行や空白があっても blob は連続化（IsUnbrokenSenryuCandidate は raw に改行があると偽）
	got := NormalizeSenryuBlob("  あ い う \t えお  ")
	if got != "あいうえお" {
		t.Fatalf("got %q", got)
	}
}

func TestSplitBlobBySentenceDelimiters(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		blob string
		want []string
	}{
		{name: "empty", blob: "", want: nil},
		{name: "no_delimiter", blob: "あいうえお", want: []string{"あいうえお"}},
		{name: "period", blob: "あ。いう", want: []string{"あ", "いう"}},
		{name: "fullwidth_period", blob: "あ．いう", want: []string{"あ", "いう"}},
		{name: "question_marks", blob: "え?お", want: []string{"え", "お"}},
		{name: "fullwidth_question", blob: "え？お", want: []string{"え", "お"}},
		{name: "exclamation", blob: "か!き", want: []string{"か", "き"}},
		{name: "fullwidth_exclamation", blob: "か！き", want: []string{"か", "き"}},
		{name: "leading_delimiter", blob: "。あい", want: []string{"あい"}},
		{name: "trailing_delimiter", blob: "あい。", want: []string{"あい"}},
		{name: "consecutive_delimiters", blob: "あ。。い", want: []string{"あ", "い"}},
		{name: "delimiters_only", blob: "。。", want: nil},
		{name: "mixed", blob: "あ？い。う", want: []string{"あ", "い", "う"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SplitBlobBySentenceDelimiters(tt.blob)
			if len(got) != len(tt.want) {
				t.Fatalf("len=%d want %d got %#v want %#v", len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("got %#v want %#v", got, tt.want)
				}
			}
		})
	}
}

func TestIsUnbrokenSenryuCandidate(t *testing.T) {
	t.Parallel()
	// 17 ひらがな（ふるいけや + かわずとびこむ + みずのおと）
	haiku := "ふるいけやかわずとびこむみずのおと"
	blob, ok := IsUnbrokenSenryuCandidate(haiku)
	if !ok || blob != haiku {
		t.Fatalf("ok=%v blob=%q", ok, blob)
	}
	_, ok = IsUnbrokenSenryuCandidate("a\nb")
	if ok {
		t.Fatal("newline should reject")
	}
	_, ok = IsUnbrokenSenryuCandidate("short")
	if ok {
		t.Fatal("too short")
	}
	long := strings.Repeat("あ", unbrokenSenryuMaxRunes+1)
	_, ok = IsUnbrokenSenryuCandidate(long)
	if ok {
		t.Fatal("too long")
	}
	// スペース区切りでも blob は17文字
	blob, ok = IsUnbrokenSenryuCandidate("ふるいけや かわずとびこむ みずのおと")
	if !ok || blob != haiku {
		t.Fatalf("ok=%v blob=%q want %q", ok, blob, haiku)
	}
}

func TestFormatSenryuReply_withPercentS(t *testing.T) {
	t.Parallel()
	got := FormatSenryuReply("前: %s 後", "みどり")
	want := "前: みどり 後"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFormatSenryuReply_fallbackQuote(t *testing.T) {
	t.Parallel()
	got := FormatSenryuReply("固定のみ", "あいう")
	want := "固定のみ\n「あいう」"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFormatSenryuReply_twoPlaceholdersUsesFallback(t *testing.T) {
	t.Parallel()
	got := FormatSenryuReply("%s と %s", "あ")
	want := "%s と %s\n「あ」"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
