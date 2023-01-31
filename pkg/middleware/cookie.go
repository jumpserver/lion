package middleware

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"lion/pkg/config"
	"lion/pkg/jms-sdk-go/model"
	"lion/pkg/jms-sdk-go/service"
	"lion/pkg/logger"
)

func JmsCookieAuth(jmsService *service.JMService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			err  error
			user *model.User
		)
		reqCookies := ctx.Request.Cookies()
		var cookies = make(map[string]string)
		for _, cookie := range reqCookies {
			cookies[cookie.Name] = cookie.Value
		}
		user, err = jmsService.CheckUserCookie(cookies)
		if err != nil {
			logger.Errorf("Check user cookie failed: %+v %s", cookies, err.Error())
			loginUrl := fmt.Sprintf("/core/auth/login/?next=%s", url.QueryEscape(ctx.Request.URL.RequestURI()))
			ctx.Redirect(http.StatusFound, loginUrl)
			ctx.Abort()
			return
		}
		ctx.Set(config.GinCtxUserKey, user)
	}
}

func HTTPMiddleDebugAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		switch ctx.ClientIP() {
		case "127.0.0.1", "localhost":
			return
		default:
			_ = ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid host %s", ctx.ClientIP()))
			return
		}
	}
}
