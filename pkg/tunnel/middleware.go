package tunnel

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"guacamole-client-go/pkg/jms-sdk-go/model"
)

func (g *GuacamoleTunnelServer) SessionAuth() gin.HandlerFunc {
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
		fmt.Println(cookies)
		user, err = g.JmsService.GetUserById("68f1648b-5c6c-4f47-97a1-c47c192458e3")
		if err != nil {
			fmt.Println(err)
			return
		}
		// TODO: 校验API
		//user, err = g.JmsService.CheckUserCookie(cookies)
		//if err != nil {
		//	logger.Debug(err.Error())
		//	loginUrl := fmt.Sprintf("/core/auth/login/?next=%s", url.QueryEscape(ctx.Request.URL.RequestURI()))
		//	ctx.Redirect(http.StatusFound, loginUrl)
		//	ctx.Abort()
		//	return
		//}
		ctx.Set(ginCtxUserKey, user)
	}
}
