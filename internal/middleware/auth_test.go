package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "0123456789abcdef0123456789abcdef"

// TestAuthRequiredAcceptsGeneratedToken 验证签发和校验路径使用相同的 HS256/issuer 约束。
func TestAuthRequiredAcceptsGeneratedToken(t *testing.T) {
	token, err := GenerateToken(testJWTSecret, 1, "admin")
	if err != nil {
		t.Fatalf("生成 JWT 失败: %v", err)
	}
	if status := requestWithToken(token); status != http.StatusOK {
		t.Fatalf("有效 JWT 返回状态 %d，期望 200", status)
	}
}

// TestAuthRequiredRejectsWrongAlgorithmAndIssuer 验证错误算法和 issuer 的 Token 均被拒绝。
func TestAuthRequiredRejectsWrongAlgorithmAndIssuer(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name   string
		method jwt.SigningMethod
		issuer string
	}{
		{name: "错误算法", method: jwt.SigningMethodHS384, issuer: tokenIssuer},
		{name: "错误签发方", method: jwt.SigningMethodHS256, issuer: "other"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			claims := Claims{
				UserID:   1,
				Username: "admin",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
					IssuedAt:  jwt.NewNumericDate(now),
					Issuer:    test.issuer,
				},
			}
			token, err := jwt.NewWithClaims(test.method, claims).SignedString([]byte(testJWTSecret))
			if err != nil {
				t.Fatalf("签发测试 JWT 失败: %v", err)
			}
			if status := requestWithToken(token); status != http.StatusUnauthorized {
				t.Fatalf("非法 JWT 返回状态 %d，期望 401", status)
			}
		})
	}
}

// requestWithToken 通过最小 Gin 路由执行一次认证请求并返回状态码。
func requestWithToken(token string) int {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthRequired(testJWTSecret))
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response.Code
}
