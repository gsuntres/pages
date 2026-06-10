package pages

import (
	"os"
	"log"
	"testing"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"

	"git.gsuntres.com/gsuntres/pkg/commons"
)

func setupRouterDefaultMode() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	root, err := PageParse("./.test/root")
	if err != nil {
		log.Fatalf("Failed to parse pages %v", err)
	}

	Bootstrap(r, root)

	instance := NewPages()
	instance.AddTemplatesFromGroup(root)

	r.HTMLRender = instance

	return r
}

func TestRootIndex(t *testing.T) {
	w := httptest.NewRecorder()

	r := setupRouterDefaultMode()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	r.ServeHTTP(w, req)

	// 5. Assert the HTTP Status Code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 6. Assert the response body
	response := string(w.Body.Bytes())
	
	expectedBytes, _ := os.ReadFile("./.test/root_rendering_response.html")
	expectedBody := string(expectedBytes)

	expected := commons.StringNormalize(expectedBody)
	actual := commons.StringNormalize(response)

	if actual != expected {
		t.Error("Unexpected rendered html")
	}
}

func TestRoute(t *testing.T) {
	w := httptest.NewRecorder()

	r := setupRouterDefaultMode()

	req, err := http.NewRequest("GET", "/group1", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	r.ServeHTTP(w, req)

	// 5. Assert the HTTP Status Code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 6. Assert the response body
	response := string(w.Body.Bytes())

	expectedBytes, _ := os.ReadFile("./.test/group1_rendering_index.html")
	expectedBody := string(expectedBytes)

	expected := commons.StringNormalize(expectedBody)
	actual := commons.StringNormalize(response)

	if actual != expected {
		t.Error("Unexpected rendered html")
	}
}

func TestRouteWithId(t *testing.T) {
	w := httptest.NewRecorder()

	r := setupRouterDefaultMode()

	req, err := http.NewRequest("GET", "/group1/1", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	r.ServeHTTP(w, req)

	// 5. Assert the HTTP Status Code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 6. Assert the response body
	response := string(w.Body.Bytes())

	expectedBytes, _ := os.ReadFile("./.test/group1_id_rendering.html")
	expectedBody := string(expectedBytes)

	expected := commons.StringNormalize(expectedBody)
	actual := commons.StringNormalize(response)
	
	if actual != expected {
		t.Error("Unexpected rendered html")
	}
}

func TestDeepRoute(t *testing.T) {
	w := httptest.NewRecorder()

	r := setupRouterDefaultMode()

	req, err := http.NewRequest("GET", "/l1/l2/l3", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	r.ServeHTTP(w, req)

	// 5. Assert the HTTP Status Code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 6. Assert the response body
	response := string(w.Body.Bytes())
	
	expectedBytes, _ := os.ReadFile("./.test/group_deep_routing.html")
	expectedBody := string(expectedBytes)

	expected := commons.StringNormalize(expectedBody)
	actual := commons.StringNormalize(response)

	if actual != expected {
		t.Error("Unexpected rendered html")
	}
}

func setupRouterLocal() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	instance := NewPagesWithProps(&PagesProps{
		Mode: ModeLocal,
	})

	r.HTMLRender = instance

	return r
}
