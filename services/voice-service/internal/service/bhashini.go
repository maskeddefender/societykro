// Package service implements the voice-service business logic and external
// API integrations.
package service

import (
	"context"
	"fmt"

	"github.com/societykro/go-common/logger"
	"github.com/societykro/voice-service/internal/model"
)

// BhashiniClient wraps calls to the Bhashini ULCA pipeline APIs for ASR
// (automatic speech recognition), NMT (neural machine translation) and LID
// (language identification).
//
// Real Bhashini API flow:
//  1. POST to pipelineURL with JSON body:
//     {
//       "pipelineTasks": [{
//         "taskType": "asr" | "translation" | "tts",
//         "config": {
//           "language": { "sourceLanguage": "<code>" },
//           "serviceId": "<service-id-from-pipeline-config>"
//         }
//       }],
//       "inputData": {
//         "audio": [{ "audioContent": "<base64>" }]   // for ASR
//         "input":  [{ "source": "<text>" }]           // for NMT/LID
//       }
//     }
//  2. Parse JSON response for output text / translation / detected language.
//  3. Authorization via headers: userID and ulcaApiKey.
type BhashiniClient struct {
	userID      string
	apiKey      string
	pipelineURL string
}

// NewBhashiniClient creates a BhashiniClient from the provided credentials.
// Typically sourced from BHASHINI_USER_ID, BHASHINI_API_KEY and
// BHASHINI_PIPELINE_URL environment variables.
func NewBhashiniClient(userID, apiKey, pipelineURL string) *BhashiniClient {
	return &BhashiniClient{
		userID:      userID,
		apiKey:      apiKey,
		pipelineURL: pipelineURL,
	}
}

// Transcribe sends audio (base64-encoded) to the Bhashini ASR pipeline and
// returns the recognised text.
//
// STUB: Returns a placeholder response. Replace with a real HTTP call to the
// Bhashini compute endpoint once service-IDs are configured.
func (b *BhashiniClient) Transcribe(ctx context.Context, audioBase64, sourceLanguage string) (*model.TranscribeResponse, error) {
	logger.Log.Info().
		Str("source_language", sourceLanguage).
		Int("audio_bytes", len(audioBase64)).
		Msg("Bhashini ASR called (stub)")

	// Bhashini API integration pending -- returning stub response.
	return &model.TranscribeResponse{
		OriginalText:     fmt.Sprintf("[stub] Transcribed text for language %s", sourceLanguage),
		OriginalLanguage: sourceLanguage,
		Confidence:       0.85,
	}, nil
}

// Translate sends text through the Bhashini NMT pipeline.
//
// STUB: Prepends "[translated] " to the input text.
func (b *BhashiniClient) Translate(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
	logger.Log.Info().
		Str("source", sourceLang).
		Str("target", targetLang).
		Int("text_len", len(text)).
		Msg("Bhashini NMT called (stub)")

	// Bhashini API integration pending -- returning stub response.
	return "[translated] " + text, nil
}

// DetectLanguage identifies the language of the given text using Bhashini LID.
//
// STUB: Always returns Hindi ("hi") with 0.95 confidence.
func (b *BhashiniClient) DetectLanguage(ctx context.Context, text string) (string, float64, error) {
	logger.Log.Info().
		Int("text_len", len(text)).
		Msg("Bhashini LID called (stub)")

	// Bhashini API integration pending -- returning stub response.
	return "hi", 0.95, nil
}
