package middleware

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"guacamole-client-go/pkg/config"
	"guacamole-client-go/pkg/jms-sdk-go/model"
	"guacamole-client-go/pkg/jms-sdk-go/service"
	"guacamole-client-go/pkg/logger"
)

func SessionAuth(jmsService *service.JMService) gin.HandlerFunc {
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
