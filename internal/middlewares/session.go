package middlewares

import (
	"context"
	"interview/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const cookieName = "ice_session_id"

func SessionMiddleware(logger log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sessionId string
		cookie, err := c.Request.Cookie(cookieName)
		if err != nil {
			sessionId = uuid.New().String()
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
