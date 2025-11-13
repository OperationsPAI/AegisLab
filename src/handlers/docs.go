package handlers

import "github.com/gin-gonic/gin"

// SwaggerModelsDoc is a documentation-only endpoint that ensures all DTO models are included in Swagger.
// This endpoint should NEVER be registered in the actual router.
//
//	@Summary		API Model Definitions
//	@Description	Virtual endpoint for including all DTO type definitions in Swagger documentation. DO NOT USE in production.
//	@Tags			Documentation
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.StreamEvent			"Server-Sent Event payload structure"
//	@Success		200	{object}	dto.DatapackResult		"Datapack result structure"
//	@Success		200	{object}	dto.ExecutionResult		"Execution result structure"
//	@Success		200	{object}	dto.InfoPayloadTemplate	"Information payload template"
//	@Success		200	{object}	dto.JobMessage			"k8s Job message structure"
//	@Success		200	{object}	consts.SSEEventName		"SSE event name constants"
//	@Router			/api/_docs/models [get]
func SwaggerModelsDoc(c *gin.Context) {}
