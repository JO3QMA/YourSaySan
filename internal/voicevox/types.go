package voicevox

// Speaker はVoiceVox Engine APIの話者情報
type Speaker struct {
	Name        string  `json:"name"`
	SpeakerUUID string  `json:"speaker_uuid"`
	Styles      []Style `json:"styles"`
}

// Style は話者のスタイル情報
type Style struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// AudioQuery は音声クエリ情報
type AudioQuery struct {
	AccentPhrases      []AccentPhrase `json:"accent_phrases"`
	SpeedScale         float64        `json:"speedScale"`
	PitchScale         float64        `json:"pitchScale"`
	IntonationScale    float64        `json:"intonationScale"`
	VolumeScale        float64        `json:"volumeScale"`
	PrePhonemeLength   float64        `json:"prePhonemeLength"`
	PostPhonemeLength  float64        `json:"postPhonemeLength"`
	OutputSamplingRate int            `json:"outputSamplingRate"`
	OutputStereo       bool           `json:"outputStereo"`
	Kana               string         `json:"kana,omitempty"`
}

// AccentPhrase はアクセント句情報
type AccentPhrase struct {
	Moras            []Mora `json:"moras"`
	Accent           int    `json:"accent"`
	PauseMora        *Mora  `json:"pauseMora,omitempty"`
	IsInterrogative  bool   `json:"isInterrogative"`
}

// Mora はモーラ情報
type Mora struct {
	Text            string   `json:"text"`
	Consonant       *string  `json:"consonant,omitempty"`
	ConsonantLength *float64 `json:"consonantLength,omitempty"`
	Vowel           string   `json:"vowel"`
	VowelLength     float64  `json:"vowelLength"`
	Pitch           float64  `json:"pitch"`
}

