package app

import (
	"lifeline_backend/internal/config"
	wshub "lifeline_backend/internal/websocket"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// App is the central dependency container.
// All dependencies are injected here and passed to routes/handlers.
type App struct {
	DB     *gorm.DB
	Logger *zap.Logger
	Router *gin.Engine
	Config *config.Config
	Hub    *wshub.Hub
}
