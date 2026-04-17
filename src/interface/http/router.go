package httpinterface

import (
	"aegis/middleware"
	"aegis/router"

	"github.com/gin-gonic/gin"
)

func NewGinEngine(handlers *router.Handlers, middlewareService middleware.Service) *gin.Engine {
	return router.New(handlers, middlewareService)
}
