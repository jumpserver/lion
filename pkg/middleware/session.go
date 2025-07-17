package middleware

import (
	"net/http"

	ginSessions "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"lion/pkg/config"
	"lion/pkg/logger"

	"github.com/jumpserver-dev/sdk-go/service"
)

func GinSessionAuth(store ginSessions.Store) gin.HandlerFunc {
	return ginSessions.Sessions(config.GinSessionName, store)
}

func SessionAuth(jmsService *service.JMService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ginSession := ginSessions.Default(ctx)
		if result := ginSession.Get(config.GinSessionKey); result != nil {
			logger.Errorf("Token auth failed %+v", ginSession)
			if uid, ok := result.(string); ok {
				if user, err := jmsService.GetUserById(uid); err == nil {
					ctx.Set(config.GinCtxUserKey, user)
					logger.Debugf("Token auth user: %s", user)
					return
				}
			}
		}
		logger.Errorf("Token auth failed %+v", ginSession)
		ctx.Status(http.StatusForbidden)
		ctx.Abort()
	}
}
