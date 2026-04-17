package service

import (
	"context"
	"fmt"

	"github.com/societykro/voice-service/internal/model"
)

// VoiceService orchestrates transcription, translation and language detection
// by delegating to the BhashiniClient.
type VoiceService struct {
	bhashini *BhashiniClient
}

// NewVoiceService creates a VoiceService backed by the given BhashiniClient.
func NewVoiceService(bhashini *BhashiniClient) *VoiceService {
	return &VoiceService{bhashini: bhashini}
}

// Transcribe converts audio to text. When the source language is not provided
// it is auto-detected. If the original language is not English, an English
// translation is included in the response.
func (s *VoiceService) Transcribe(ctx context.Context, req model.TranscribeRequest) (*model.TranscribeResponse, error) {
	if req.AudioBase64 == "" && req.AudioURL == "" {
		return nil, fmt.Errorf("either audio_url or audio_base64 is required")
	}

	sourceLang := req.SourceLanguage

	// Auto-detect language when the caller did not specify one.
	if sourceLang == "" {
		detected, _, err := s.bhashini.DetectLanguage(ctx, "audio-content")
		if err != nil {
			return nil, fmt.Errorf("language detection failed: %w", err)
		}
		sourceLang = detected
	}

	audio := req.AudioBase64
	if audio == "" {
		// TODO: download audio from AudioURL and base64-encode it.
		audio = req.AudioURL
	}

	resp, err := s.bhashini.Transcribe(ctx, audio, sourceLang)
	if err != nil {
		return nil, fmt.Errorf("transcription failed: %w", err)
	}

	resp.OriginalLanguage = sourceLang

	// Translate to English when the source language differs.
	if sourceLang != "en" {
		translated, err := s.bhashini.Translate(ctx, resp.OriginalText, sourceLang, "en")
		if err != nil {
			return nil, fmt.Errorf("translation to English failed: %w", err)
		}
		resp.EnglishText = translated
	} else {
		resp.EnglishText = resp.OriginalText
	}

	return resp, nil
}

// Translate translates text between two Bhashini-supported languages.
func (s *VoiceService) Translate(ctx context.Context, req model.TranslateRequest) (*model.TranslateResponse, error) {
	if req.Text == "" {
		return nil, fmt.Errorf("text is required")
	}
	if req.SourceLanguage == "" || req.TargetLanguage == "" {
		return nil, fmt.Errorf("source_language and target_language are required")
	}

	translated, err := s.bhashini.Translate(ctx, req.Text, req.SourceLanguage, req.TargetLanguage)
	if err != nil {
		return nil, fmt.Errorf("translation failed: %w", err)
	}

	return &model.TranslateResponse{
		TranslatedText: translated,
		SourceLanguage: req.SourceLanguage,
		TargetLanguage: req.TargetLanguage,
	}, nil
}

// DetectLanguage identifies the language of the provided text.
func (s *VoiceService) DetectLanguage(ctx context.Context, req model.DetectLanguageRequest) (*model.DetectLanguageResponse, error) {
	if req.Text == "" {
		return nil, fmt.Errorf("text is required")
	}

	lang, conf, err := s.bhashini.DetectLanguage(ctx, req.Text)
	if err != nil {
		return nil, fmt.Errorf("language detection failed: %w", err)
	}

	return &model.DetectLanguageResponse{
		Language:   lang,
		Confidence: conf,
	}, nil
}

// GetSupportedLanguages returns the 22 scheduled Indian languages and their
// Bhashini pipeline availability.
func (s *VoiceService) GetSupportedLanguages() []model.SupportedLanguage {
	return []model.SupportedLanguage{
		{Code: "as", Name: "Assamese", NativeName: "\u0985\u09b8\u09ae\u09c0\u09af\u09bc\u09be", BhashiniASR: true, BhashiniNMT: true},
		{Code: "bn", Name: "Bengali", NativeName: "\u09ac\u09be\u0982\u09b2\u09be", BhashiniASR: true, BhashiniNMT: true},
		{Code: "brx", Name: "Bodo", NativeName: "\u092c\u0930\u0952", BhashiniASR: false, BhashiniNMT: true},
		{Code: "doi", Name: "Dogri", NativeName: "\u0921\u094b\u0917\u0930\u0940", BhashiniASR: false, BhashiniNMT: true},
		{Code: "en", Name: "English", NativeName: "English", BhashiniASR: true, BhashiniNMT: true},
		{Code: "gu", Name: "Gujarati", NativeName: "\u0a97\u0ac1\u0a9c\u0ab0\u0abe\u0aa4\u0ac0", BhashiniASR: true, BhashiniNMT: true},
		{Code: "hi", Name: "Hindi", NativeName: "\u0939\u093f\u0928\u094d\u0926\u0940", BhashiniASR: true, BhashiniNMT: true},
		{Code: "kn", Name: "Kannada", NativeName: "\u0c95\u0ca8\u0ccd\u0ca8\u0ca1", BhashiniASR: true, BhashiniNMT: true},
		{Code: "ks", Name: "Kashmiri", NativeName: "\u0915\u0949\u0936\u0941\u0930", BhashiniASR: false, BhashiniNMT: true},
		{Code: "gom", Name: "Konkani", NativeName: "\u0915\u094b\u0902\u0915\u0923\u0940", BhashiniASR: false, BhashiniNMT: true},
		{Code: "mai", Name: "Maithili", NativeName: "\u092e\u0948\u0925\u093f\u0932\u0940", BhashiniASR: false, BhashiniNMT: true},
		{Code: "ml", Name: "Malayalam", NativeName: "\u0d2e\u0d32\u0d2f\u0d3e\u0d33\u0d02", BhashiniASR: true, BhashiniNMT: true},
		{Code: "mni", Name: "Manipuri", NativeName: "\u09ae\u09c8\u09a4\u09c8\u09b2\u09cb\u09a8", BhashiniASR: false, BhashiniNMT: true},
		{Code: "mr", Name: "Marathi", NativeName: "\u092e\u0930\u093e\u0920\u0940", BhashiniASR: true, BhashiniNMT: true},
		{Code: "ne", Name: "Nepali", NativeName: "\u0928\u0947\u092a\u093e\u0932\u0940", BhashiniASR: true, BhashiniNMT: true},
		{Code: "or", Name: "Odia", NativeName: "\u0b13\u0b21\u0b3c\u0b3f\u0b06", BhashiniASR: true, BhashiniNMT: true},
		{Code: "pa", Name: "Punjabi", NativeName: "\u0a2a\u0a70\u0a1c\u0a3e\u0a2c\u0a40", BhashiniASR: true, BhashiniNMT: true},
		{Code: "sa", Name: "Sanskrit", NativeName: "\u0938\u0902\u0938\u094d\u0915\u0943\u0924\u092e\u094d", BhashiniASR: false, BhashiniNMT: true},
		{Code: "sat", Name: "Santali", NativeName: "\u1c65\u1c5f\u1c71\u1c5b\u1c5f\u1c6e\u1c64", BhashiniASR: false, BhashiniNMT: true},
		{Code: "sd", Name: "Sindhi", NativeName: "\u0633\u0646\u068c\u064a", BhashiniASR: false, BhashiniNMT: true},
		{Code: "ta", Name: "Tamil", NativeName: "\u0ba4\u0bae\u0bbf\u0bb4\u0bcd", BhashiniASR: true, BhashiniNMT: true},
		{Code: "te", Name: "Telugu", NativeName: "\u0c24\u0c46\u0c32\u0c41\u0c17\u0c41", BhashiniASR: true, BhashiniNMT: true},
		{Code: "ur", Name: "Urdu", NativeName: "\u0627\u0631\u062f\u0648", BhashiniASR: true, BhashiniNMT: true},
	}
}
