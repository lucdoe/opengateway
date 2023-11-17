package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lucdoe/capstone_gateway/internal/middlewares"
)

const MaxAgeCORS = 12 * time.Hour

func checkHeaders(t *testing.T, recorder *httptest.ResponseRecorder) {
	headers := recorder.Header()
	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "POST, GET, OPTIONS, PUT, PATCH",
		"Access-Control-Allow-Headers": "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
		"Max-Age":                      fmt.Sprintf("%f", MaxAgeCORS.Seconds()),
	}

	for key, expectedValue := range expectedHeaders {
		if value := headers.Get(key); value != expectedValue {
			t.Errorf("Header %s = %v, want %v", key, value, expectedValue)
		}
	}
}

func TestCORS(t *testing.T) {
	router := gin.Default()
	router.Use(middlewares.CORS)

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	tests := []struct {
		name       string
		method     string
		statusCode int
	}{
		{"GET Method", http.MethodGet, http.StatusOK},
		{"OPTIONS Method", http.MethodOptions, http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request, _ := http.NewRequest(tt.method, "/", nil)
			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.statusCode {
				t.Errorf("Status code = %v, want %v", recorder.Code, tt.statusCode)
			}

			checkHeaders(t, recorder)
		})
	}
}
