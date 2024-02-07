package middlewares

import (
	"fmt"
	"interview/pkg/log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
)

func TestSettingNewSessionToRequests(t *testing.T) {
	res := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(res)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	logger, _ := log.NewForTest()
	handler := SessionMiddleware(logger)
	handler(c)
	ctx := c.Request.Context()
	session := ctx.Value("SessionId")
	assert.NotNil(t, session)
}

func TestGettingSessionFromRequestCookie(t *testing.T) {
	res := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(res)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	sessionId := uuid.New().String()
	c.Request.Header.Add("Cookie", fmt.Sprintf("%s=%s", cookieName, sessionId))
	logger, _ := log.NewForTest()
	handler := SessionMiddleware(logger)
	handler(c)
	ctx := c.Request.Context()
	session := ctx.Value("SessionId")
	assert.NotNil(t, session)
	assert.Equal(t, sessionId, session)
}
