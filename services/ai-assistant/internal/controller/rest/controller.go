package rest

import (
	"go.uber.org/fx"
	"net/http"

	"soa-video-streaming/services/ai-assistant/internal/service"

	"github.com/gin-gonic/gin"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(NewAIController),
	)
}

type AIController struct {
	assistant *service.Assistant
}

func NewAIController(assistant *service.Assistant) *AIController {
	return &AIController{assistant: assistant}
}

func (c *AIController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/command", c.HandleCommand)
}

type CommandRequest struct {
	Prompt string `json:"prompt"`
}

func (c *AIController) HandleCommand(ctx *gin.Context) {
	var req CommandRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Prompt == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}

	resp, err := c.assistant.ProcessCommand(ctx.Request.Context(), req.Prompt)
	if err != nil {
		code := http.StatusInternalServerError
		if err.Error() == "token limit exceeded" || err.Error() == "potential prompt injection detected" {
			code = http.StatusBadRequest
		}

		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"response": resp})
}
