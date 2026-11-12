package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"autostack/internal/config"
	"autostack/internal/utils"
	"github.com/gin-gonic/gin"
)

func TestOrderSummaryAllowsSummaryOnlyPermission(t *testing.T) {
	const secret = "summary-route-test-secret"
	cfg := &config.Config{}
	cfg.JWT.Secret = secret
	s := &Server{config: cfg, router: gin.New()}
	s.setupRoutes()
	for _, tc := range []struct {
		name        string
		permissions []string
		want        int
	}{
		{"summary only", []string{"route:product.order_summary:read"}, http.StatusBadRequest},
		{"unrelated permission", []string{"route:product.local_products:read"}, http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			token, err := utils.GenerateTokenWithPermissions(1, 90000000001, "test", "user", tc.permissions, secret, 1)
			if err != nil {
				t.Fatal(err)
			}
			// Invalid query must reach the summary handler and fail validation without touching a database.
			request := httptest.NewRequest(http.MethodGet, "/api/v1/order/stats/summary?auth_id=invalid", nil)
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			s.router.ServeHTTP(response, request)
			if response.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", response.Code, tc.want, response.Body.String())
			}
		})
	}
}
