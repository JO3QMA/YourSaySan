package senryu

import "testing"

func TestCountMoraeInReading(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"アイウ", 3},
		{"キャ", 1},
		{"キッタ", 3},
		{"コーン", 3},
		{"かきくけこ", 5},
	}
	for _, tt := range tests {
		if got := CountMoraeInReading(tt.in); got != tt.want {
			t.Errorf("CountMoraeInReading(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}
