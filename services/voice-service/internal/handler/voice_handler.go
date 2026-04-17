// Package handler provides HTTP handlers for the voice service endpoints.
package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/societykro/go-common/middleware"
	"github.com/societykro/go-common/response"
	"github.com/societykro/voice-service/internal/model"
	"github.com/societykro/voice-service/internal/service"
)

// VoiceHandler exposes voice-service functionality over HTTP.
type VoiceHandler struct {
	svc *service.VoiceService
}

// NewVoiceHandler creates a VoiceHandler backed by the given VoiceService.
func NewVoiceHandler(svc *service.VoiceService) *VoiceHandler {
	return &VoiceHandler{svc: svc}
}

// Transcribe handles POST /voice/transcribe.
// It accepts audio (base64 or URL) and returns the transcription along with
// an English translation when the source language is not English.
func (h *VoiceHandler) Transcribe(c *fiber.Ctx) error {
	if _, ok := middleware.GetUserID(c); !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	var req model.TranscribeRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if req.AudioBase64 == "" && req.AudioURL == "" {
		return response.BadRequest(c, "Either audio_url or audio_base64 is required")
	}

	result, err := h.svc.Transcribe(c.Context(), req)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, result)
}

// Translate handles POST /voice/translate.
// It translates text between two Bhashini-supported languages.
func (h *VoiceHandler) Translate(c *fiber.Ctx) error {
	if _, ok := middleware.GetUserID(c); !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	var req model.TranslateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if req.Text == "" {
		return response.BadRequest(c, "text is required")
	}
	if req.SourceLanguage == "" || req.TargetLanguage == "" {
		return response.BadRequest(c, "source_language and target_language are required")
	}

	result, err := h.svc.Translate(c.Context(), req)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, result)
}

// DetectLanguage handles POST /voice/detect-language.
// It identifies the language of the provided text.
func (h *VoiceHandler) DetectLanguage(c *fiber.Ctx) error {
	if _, ok := middleware.GetUserID(c); !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	var req model.DetectLanguageRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if req.Text == "" {
		return response.BadRequest(c, "text is required")
	}

	result, err := h.svc.DetectLanguage(c.Context(), req)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, result)
}

// GetSupportedLanguages handles GET /languages.
// This endpoint is public (no authentication required) and returns all 22
// scheduled Indian languages with their Bhashini pipeline availability.
func (h *VoiceHandler) GetSupportedLanguages(c *fiber.Ctx) error {
	langs := h.svc.GetSupportedLanguages()
	return response.OK(c, langs)
}
