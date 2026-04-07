package voicevox

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient は httptest.Server のURLを使うClientを作成するヘルパー。
// リトライのバックオフをテスト用に最小化する。
func newTestClient(serverURL string) *Client {
	c := NewClient(serverURL)
	c.retryBackoff = 1 * time.Millisecond
	c.retryBackoffMax = 5 * time.Millisecond
	return c
}

// --- HTTPError テスト ---

func TestHTTPError_Error(t *testing.T) {
	err := &HTTPError{StatusCode: 422, Message: "invalid speaker"}
	assert.Equal(t, "HTTP 422: invalid speaker", err.Error())
}

// --- GetSpeakers テスト ---

func TestClient_GetSpeakers_Success(t *testing.T) {
	speakers := []Speaker{
		{Name: "四国めたん", SpeakerUUID: "uuid-1", Styles: []Style{{Name: "ノーマル", ID: 2}}},
		{Name: "ずんだもん", SpeakerUUID: "uuid-2", Styles: []Style{{Name: "ノーマル", ID: 3}}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/speakers", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(speakers))
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	got, err := client.GetSpeakers(context.Background())
	require.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "四国めたん", got[0].Name)
	assert.Equal(t, 2, got[0].Styles[0].ID)
}

func TestClient_GetSpeakers_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.GetSpeakers(context.Background())
	assert.Error(t, err)

	var httpErr *HTTPError
	assert.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusInternalServerError, httpErr.StatusCode)
}

func TestClient_GetSpeakers_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("not valid json"))
		require.NoError(t, err)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.GetSpeakers(context.Background())
	assert.Error(t, err)
}

// --- Speak テスト ---

func TestClient_Speak_Success(t *testing.T) {
	audioQueryResponse := AudioQuery{
		AccentPhrases:   []AccentPhrase{},
		SpeedScale:      1.0,
		VolumeScale:     1.0,
		OutputSamplingRate: 24000,
	}
	expectedAudio := []byte("fake-wav-data")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/audio_query":
			assert.Equal(t, "1", r.URL.Query().Get("speaker"))
			assert.NotEmpty(t, r.URL.Query().Get("text"))
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(audioQueryResponse))

		case r.Method == "POST" && r.URL.Path == "/synthesis":
			assert.Equal(t, "1", r.URL.Query().Get("speaker"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			// synthesisはAudioQueryを受け取りWAVを返す
			var q AudioQuery
			require.NoError(t, json.NewDecoder(r.Body).Decode(&q))
			// サンプリングレートが48kHzに設定されていることを確認
			assert.Equal(t, 48000, q.OutputSamplingRate)
			w.WriteHeader(http.StatusOK)
			_, werr := w.Write(expectedAudio)
			require.NoError(t, werr)

		default:
			http.Error(w, "unexpected path: "+r.URL.Path, http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	audio, err := client.Speak(context.Background(), "こんにちは", 1)
	require.NoError(t, err)
	assert.Equal(t, expectedAudio, audio)
}

func TestClient_Speak_AudioQueryError_4xx_NoRetry(t *testing.T) {
	var callCount int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		// 422 Unprocessable Entity（不正な話者ID）
		http.Error(w, "invalid speaker", http.StatusUnprocessableEntity)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.Speak(context.Background(), "テスト", 9999)
	assert.Error(t, err)

	var httpErr *HTTPError
	assert.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusUnprocessableEntity, httpErr.StatusCode)

	// 4xxはリトライしない（1回のみ）
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))
}

func TestClient_Speak_ServerError_RetriesAndFails(t *testing.T) {
	var callCount int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.Speak(context.Background(), "テスト", 1)
	assert.Error(t, err)

	// maxRetries=3 回試みる
	assert.Equal(t, int32(client.maxRetries), atomic.LoadInt32(&callCount))
}

func TestClient_Speak_RetrySucceedsOnSecondAttempt(t *testing.T) {
	var callCount int32

	audioQueryResponse := AudioQuery{OutputSamplingRate: 24000}
	expectedAudio := []byte("audio-data")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&callCount, 1)

		if r.URL.Path == "/audio_query" {
			if n == 1 {
				// 1回目は503
				http.Error(w, "service unavailable", http.StatusServiceUnavailable)
				return
			}
			// 2回目は成功
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(audioQueryResponse))
			return
		}

		if r.URL.Path == "/synthesis" {
			_, werr := w.Write(expectedAudio)
			require.NoError(t, werr)
		}
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	audio, err := client.Speak(context.Background(), "テスト", 1)
	require.NoError(t, err)
	assert.Equal(t, expectedAudio, audio)
}

func TestClient_Speak_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// リクエストの処理を意図的に遅延
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.Speak(ctx, "テスト", 1)
	assert.Error(t, err)
}

func TestClient_Speak_SetsOutputSamplingRateTo48kHz(t *testing.T) {
	var receivedQuery AudioQuery

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/audio_query":
			// 元のサンプリングレートは24kHz
			resp := AudioQuery{OutputSamplingRate: 24000}
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(resp))
		case "/synthesis":
			// synthesisに送られたQueryを確認
			require.NoError(t, json.NewDecoder(r.Body).Decode(&receivedQuery))
			_, werr := w.Write([]byte("audio"))
			require.NoError(t, werr)
		}
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.Speak(context.Background(), "テスト", 1)
	require.NoError(t, err)

	// クライアントが48kHzに書き換えていることを確認
	assert.Equal(t, 48000, receivedQuery.OutputSamplingRate)
}

func TestClient_FindSenryuMatch_RuneGate(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	client := newTestClient(srv.URL)
	ctx := context.Background()

	_, ok, err := client.FindSenryuMatch(ctx, "short", 1, 12, 100)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Equal(t, int32(0), calls.Load())

	long := strings.Repeat("あ", 200)
	_, ok2, err2 := client.FindSenryuMatch(ctx, long, 1, 12, 100)
	require.NoError(t, err2)
	assert.False(t, ok2)
	assert.Equal(t, int32(0), calls.Load())
}

func TestClient_FindSenryuMatch_UsesAudioQuery(t *testing.T) {
	moras := make([]map[string]any, 18)
	wantPrefix := ""
	for i := 0; i < 18; i++ {
		ch := string(rune('A' + i))
		moras[i] = map[string]any{"text": ch, "vowel": "a", "vowel_length": 0.1, "pitch": 0}
		if i < 17 {
			wantPrefix += ch
		}
	}
	raw, err := json.Marshal(map[string]any{
		"accent_phrases": []map[string]any{{
			"moras": moras, "accent": 1, "isInterrogative": false,
		}},
		"speedScale": 1, "pitchScale": 0, "intonationScale": 1, "volumeScale": 1,
		"prePhonemeLength": 0.1, "postPhonemeLength": 0.1,
		"outputSamplingRate": 24000, "outputStereo": false,
	})
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/audio_query" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(raw)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	client := newTestClient(srv.URL)
	match, ok, err := client.FindSenryuMatch(context.Background(), strings.Repeat("あ", 12), 2, 12, 100)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, wantPrefix, match)
}
