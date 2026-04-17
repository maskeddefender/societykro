// Package model defines request/response types for the voice service.
package model

// TranscribeRequest holds parameters for speech-to-text transcription.
// Provide either AudioURL or AudioBase64. If SourceLanguage is empty the
// service will attempt automatic language detection via Bhashini LID.
type TranscribeRequest struct {
	AudioURL       string `json:"audio_url,omitempty"`
	AudioBase64    string `json:"audio_base64,omitempty"`
	SourceLanguage string `json:"source_language,omitempty"` // ISO 639-1 code, e.g. "hi", "ta"
}

// TranscribeResponse contains both the original transcription and its
// English translation (if the source language is not English).
type TranscribeResponse struct {
	OriginalText     string  `json:"original_text"`
	OriginalLanguage string  `json:"original_language"`
	EnglishText      string  `json:"english_text"`
	Confidence       float64 `json:"confidence"`
}

// TranslateRequest holds parameters for text translation between two
// Bhashini-supported languages.
type TranslateRequest struct {
	Text           string `json:"text"`
	SourceLanguage string `json:"source_language"` // ISO 639-1
	TargetLanguage string `json:"target_language"` // ISO 639-1
}

// TranslateResponse contains the translated text along with the resolved
// source and target language codes.
type TranslateResponse struct {
	TranslatedText string `json:"translated_text"`
	SourceLanguage string `json:"source_language"`
	TargetLanguage string `json:"target_language"`
}

// DetectLanguageRequest holds input text for language identification.
type DetectLanguageRequest struct {
	Text string `json:"text"`
}

// DetectLanguageResponse contains the detected language and a confidence score.
type DetectLanguageResponse struct {
	Language   string  `json:"language"`
	Confidence float64 `json:"confidence"`
}

// SupportedLanguage describes one of the 22 scheduled Indian languages and
// indicates which Bhashini pipeline tasks are available for it.
type SupportedLanguage struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	NativeName  string `json:"native_name"`
	BhashiniASR bool   `json:"bhashini_asr"`
	BhashiniNMT bool   `json:"bhashini_nmt"`
}
