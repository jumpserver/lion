package middleware

import (
	"net/http"

	ginSessions "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"lion/pkg/config"
	"lion/pkg/logger"
	"lion/pkg/session"
)

func GinSessionAuth(store ginSessions.Store) gin.HandlerFunc {
	return ginSessions.Sessions(config.GinSessionName, store)
}

func SessionAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ginSession := ginSessions.Default(ctx)
		if result := ginSession.Get(config.GinSessionKey); result != nil {
			if tokenSession, ok := result.(*session.TunnelSession); ok {
				logger.Debug("token auth success ")
				ctx.Set(config.GinCtxUserKey, tokenSession.User)
				return
			}
		}
		logger.Debug("token auth failed")
		ctx.Status(http.StatusForbidden)
		ctx.Abort()
	}
}
