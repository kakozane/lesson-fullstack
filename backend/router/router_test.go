package router

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"lesson06/backend/bootstrap"
	authdto "lesson06/backend/internal/auth/dto"
	authhandler "lesson06/backend/internal/auth/handler"
	authservice "lesson06/backend/internal/auth/service"
	producthandler "lesson06/backend/internal/product/admin/handler"
	productservice "lesson06/backend/internal/product/admin/service"

	"github.com/gin-gonic/gin"
	taskhandler "lesson06/backend/internal/task/admin/handler"
	taskservice "lesson06/backend/internal/task/admin/service"
)

func TestHTTPContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sessions := authservice.NewAuthService("test-password")
	app := &bootstrap.App{Task: taskhandler.New(taskservice.New(nil)), Sessions: sessions, Auth: authhandler.New(sessions), Product: producthandler.New(productservice.NewProductService(nil))}
	r := SetupRouter(app)
	result, err := sessions.Login(authdto.LoginRequest{Username: "admin", Password: "test-password"})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, method, path, body, authorization string
		status                                  int
	}{
		{"task unauthorized", "GET", "/api/tasks", "", "", 401},
		{"invalid task", "POST", "/api/tasks", `{"title":"","version":1}`, "Bearer " + result.AccessToken, 400},
		{"invalid move", "PATCH", "/api/tasks/1/move", `{"status":"unknown","version":1}`, "Bearer " + result.AccessToken, 400},
		{"invalid task version", "DELETE", "/api/tasks/1", `{"version":0}`, "Bearer " + result.AccessToken, 400},
		{"no token", "GET", "/api/products", "", "", 401},
		{"bad token", "GET", "/api/products", "", "Bearer invalid", 401},
		{"missing bearer prefix", "GET", "/api/products", "", result.AccessToken, 401},
		{"invalid login", "POST", "/api/auth/login", `{"username":"admin","password":"wrong"}`, "", 401},
		{"trailing json", "POST", "/api/products", "{} {}", "Bearer " + result.AccessToken, 400},
		{"unknown field", "POST", "/api/products", `{"unknown":1}`, "Bearer " + result.AccessToken, 400},
		{"oversized body", "POST", "/api/products", `{"description":"` + strings.Repeat("a", 65536) + `"}`, "Bearer " + result.AccessToken, 400},
		{"invalid page", "GET", "/api/products?page=0", "", "Bearer " + result.AccessToken, 400},
		{"malformed page", "GET", "/api/products?page=abc", "", "Bearer " + result.AccessToken, 400},
		{"invalid id", "PUT", "/api/products/no", "{}", "Bearer " + result.AccessToken, 400},
		{"invalid batch", "PATCH", "/api/products/status", `{"ids":[1,1],"status":0}`, "Bearer " + result.AccessToken, 400},
		{"info", "GET", "/api/user/info", "", "Bearer " + result.AccessToken, 200},
		{"categories", "GET", "/api/categories", "", "Bearer " + result.AccessToken, 200},
		{"not found", "GET", "/api/unknown", "", "Bearer " + result.AccessToken, 404},
		{"wrong method", "POST", "/api/categories", "", "Bearer " + result.AccessToken, 405},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", tc.authorization)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("status %d body %s", w.Code, w.Body.String())
			}
			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			expectedCode := 400
			if tc.status == 200 || tc.status == 401 {
				expectedCode = tc.status
			}
			if body["code"] != float64(expectedCode) {
				t.Fatalf("code %v", body)
			}
			for _, key := range []string{"data", "message"} {
				if _, ok := body[key]; !ok {
					t.Fatalf("missing %s", key)
				}
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("missing cache control")
			}
		})
	}
	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+result.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 || sessions.Valid(result.AccessToken) {
		t.Fatal("logout failed")
	}
}
