package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"

	"github.com/gabihodoroaga/cloudrun-casbin/config"
)

// RequireAuthentication is a gin middleware that checks if the request has a valid
// jwt token issued by google
func RequireAuthentication() gin.HandlerFunc {
	clientID := config.GetConfig().ClientID
	return func(c *gin.Context) {
		logger := zap.L().With(zap.String("request_id", c.GetString("request_id")))
		authHeader := c.Request.Header.Get("Authorization")

		if authHeader == "" {
			logger.Sugar().Debugf("authorization header not found")
			c.Header("WWW-Authenticate", "Bearer realm=\"sign-in-test-app\", error=\"invalid_token\", error_description=\"Authorization header not found\"")
			c.AbortWithStatus(401)
			return
		}

		prefix := "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			logger.Sugar().Debugf("bearer prefix not found in authorization header")
			c.Header("WWW-Authenticate", "Bearer realm=\"sign-in-test-app\", error=\"invalid_token\", error_description=\"Bearer prefix not found in authorization header\"")
			c.AbortWithStatus(401)
			return
		}

		token := authHeader[strings.Index(authHeader, prefix)+len(prefix):]
		if token == "" {
			logger.Sugar().Debugf("not a valid jwt token")
			c.Header("WWW-Authenticate", "Bearer realm=\"sign-in-test-app\",error=\"invalid_token\",error_description=\"Empty jwt token\"")
			c.AbortWithStatus(401)
			return
		}

		payload, err := idtoken.Validate(c.Request.Context(), token, clientID)
		if err != nil {
			logger.Sugar().Debugf("token validation error: %s", err.Error())
			c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"sign-in-test-app\",error=\"invalid_token\",error_description=\"%s\"", err.Error()))
			c.AbortWithStatus(401)
			return
		}

		// validate IssuedAt - this is not validate by the google package
		if payload.IssuedAt == 0 || payload.IssuedAt-30 > time.Now().Unix() {
			logger.Sugar().Debugf("token validation error: Token emitted in the future")
			c.Header("WWW-Authenticate", "Bearer realm=\"sign-in-test-app\",error=\"invalid_token\",error_description=\"Token emitted in the future\"")
			c.AbortWithStatus(401)
			return
		}
		// add the user to the context
		c.Set("user", payload.Claims["email"])
	}
}
