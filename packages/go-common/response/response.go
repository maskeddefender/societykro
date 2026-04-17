package response

import "github.com/gofiber/fiber/v2"

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error"`
	Detail    string `json:"detail,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	NextCursor string      `json:"next_cursor,omitempty"`
	HasMore    bool        `json:"has_more"`
	Total      int64       `json:"total,omitempty"`
}

func OK(c *fiber.Ctx, data interface{}) error {
	return c.JSON(SuccessResponse{Success: true, Data: data})
}

func OKMessage(c *fiber.Ctx, message string) error {
	return c.JSON(SuccessResponse{Success: true, Message: message})
}

func Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{Success: true, Data: data})
}

func Paginated(c *fiber.Ctx, data interface{}, nextCursor string, hasMore bool, total int64) error {
	return c.JSON(PaginatedResponse{
		Success: true, Data: data, NextCursor: nextCursor, HasMore: hasMore, Total: total,
	})
}

func BadRequest(c *fiber.Ctx, detail string) error {
	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
		Error: "bad_request", Detail: detail,
	})
}

func Unauthorized(c *fiber.Ctx, detail string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
		Error: "unauthorized", Detail: detail,
	})
}

func Forbidden(c *fiber.Ctx, detail string) error {
	return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
		Error: "forbidden", Detail: detail,
	})
}

func NotFound(c *fiber.Ctx, detail string) error {
	return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
		Error: "not_found", Detail: detail,
	})
}

func Conflict(c *fiber.Ctx, detail string) error {
	return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
		Error: "conflict", Detail: detail,
	})
}

func InternalError(c *fiber.Ctx, detail string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Error: "internal_error", Detail: detail,
	})
}

func TooManyRequests(c *fiber.Ctx, detail string) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(ErrorResponse{
		Error: "rate_limited", Detail: detail,
	})
}
