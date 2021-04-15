package middleware

import (
	ginSessions "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"lion/pkg/config"
)

func GinSessionAuth(store ginSessions.Store) gin.HandlerFunc {
	return ginSessions.Sessions(config.GinSessionName, store)
}
