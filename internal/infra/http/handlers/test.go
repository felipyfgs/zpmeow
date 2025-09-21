package handlers

import (
	"github.com/gofiber/fiber/v2"
)

type TestHandler struct {
	*BaseHandler
}

func NewTestHandler() *TestHandler {
	return &TestHandler{
		BaseHandler: NewBaseHandler("test-handler"),
	}
}

// Test endpoint for Fiber migration
func (h *TestHandler) Test(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Fiber migration working!",
		"status":  "success",
	})
}
