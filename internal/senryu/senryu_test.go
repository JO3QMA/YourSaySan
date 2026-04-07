package senryu

import (
	"context"
	"errors"
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

type stubCounter struct {
	counts []int
	err    error
	i      int
}

func (s *stubCounter) CountMorae(_ context.Context, _ string, _ int) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	if s.i >= len(s.counts) {
		return 0, errors.New("exhausted")
	}
	n := s.counts[s.i]
	s.i++
	return n, nil
}

func TestIs575Morae(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	lines := []string{"a", "b", "c"}

	stub := &stubCounter{counts: []int{5, 7, 5}}
	ok, err := Is575Morae(ctx, stub, lines, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected true")
	}

	stub2 := &stubCounter{counts: []int{5, 6, 5}}
	ok, err = Is575Morae(ctx, stub2, lines, 1)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected false")
	}

	stubErr := &stubCounter{err: errors.New("boom")}
	_, err = Is575Morae(ctx, stubErr, lines, 1)
	if err == nil {
		t.Fatal("expected error")
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

func TestIs575MoraeUnbroken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	stub := &stubCounter{counts: []int{17}}
	ok, err := Is575MoraeUnbroken(ctx, stub, "ふるいけやかわずとびこむみずのおと", 1)
	if err != nil || !ok {
		t.Fatalf("err=%v ok=%v", err, ok)
	}
	stub2 := &stubCounter{counts: []int{16}}
	ok, err = Is575MoraeUnbroken(ctx, stub2, "x", 1)
	if err != nil || ok {
		t.Fatalf("err=%v ok=%v want false", err, ok)
	}
	stubErr := &stubCounter{err: errors.New("boom")}
	_, err = Is575MoraeUnbroken(ctx, stubErr, "x", 1)
	if err == nil {
		t.Fatal("expected error")
	}
	ok, err = Is575MoraeUnbroken(ctx, stub, "", 1)
	if err != nil || ok {
		t.Fatalf("empty blob: err=%v ok=%v", err, ok)
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
