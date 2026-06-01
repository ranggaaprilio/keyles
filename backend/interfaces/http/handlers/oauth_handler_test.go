package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestDirectPeerIPIgnoresForwardedHeaders(t *testing.T) {
	req := httptest.NewRequest("POST", "/oauth2/login", nil)
	req.RemoteAddr = "192.0.2.10:41234"
	req.Header.Set("X-Forwarded-For", "203.0.113.99")
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = req

	require.Equal(t, "192.0.2.10", directPeerIP(ctx))
}
