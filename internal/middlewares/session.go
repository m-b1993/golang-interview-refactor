package middlewares

import (
	"context"
	"interview/pkg/log"
	"time"

	"github.com/gin-gonic/gin"
)

const cookieName = "ice_session_id"

func SessionMiddleware(logger log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sessionId string
		cookie, err := c.Request.Cookie(cookieName)
		if err != nil {
			sessionId = time.Now().String()
			c.SetCookie(cookieName, sessionId, 3600, "/", "localhost", false, true)
		} else {
			sessionId = cookie.Value
		}
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "SessionId", sessionId)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
