// package server provides the Helper HTTP server, which is the service, configured with the Helper Management Service.
package httpserver

import (
	"net/http"
	"sync"

	"github.com/kubeshop/kusk-gateway/internal/helper/mocking"
)

const (
	// Server Hostname and port are not configurable since the manager needs to know them for the Envoy cluster creation
	ServerHostname           string = "127.0.0.1"
	ServerPort               uint32 = 8090
	HeaderMockID                    = "X-Kusk-Mock-ID"
	HeaderMockResponseInsert        = "X-Kusk-Mocked"
)

// HTTP Handler is the main server handler
type HTTPHandler struct {
	mockConfig *mocking.MockConfig
	mu         *sync.RWMutex
}

func (h *HTTPHandler) SetMockConfig(mockConfig *mocking.MockConfig) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.mockConfig = mockConfig
}

func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{mockConfig: mocking.NewMockConfig(), mu: &sync.RWMutex{}}
}

// ServerHTTP implements the standard net/http handler interface
func (m *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mockID := r.Header.Get(HeaderMockID)
	// Fail if no mockID header in the request
	if mockID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	mockResponse := m.mockConfig.GetMockResponse(mockID)
	// Fail if no mockID found in the MockResponses cache
	if mockResponse == nil {
		// Add Mocked header with "false" to show that we didn't find the response
		w.Header().Set(HeaderMockResponseInsert, "false")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Add Mocked header to show that we mocked the response
	w.Header().Set(HeaderMockResponseInsert, "true")
	// TODO: detect content type for the user using its request Accept Header
	mediaType := "application/json"
	data, ok := mockResponse.MediaTypeData[mediaType]
	// If no media type data (example) found - this is the simple http code in the response, write it and return
	if !ok {
		w.WriteHeader(mockResponse.StatusCode)
		return
	}
	// otherwise, set Content-Type and write the body
	w.Header().Set("Content-Type", mediaType)
	w.WriteHeader(mockResponse.StatusCode)
	w.Write(data)
}

// HealthcheckHTTPHandler handles healthcheck
type HealthcheckHTTPHandler struct {
	// Once we have everything needed to serve, set this to true
	ready bool
}

func NewHealthcheckHTTPHandler() *HealthcheckHTTPHandler {
	return &HealthcheckHTTPHandler{ready: false}
}

// Enable makes healtcheck healthy
func (h *HealthcheckHTTPHandler) Enable() {
	h.ready = true
}

// ServerHTTP implements the standard net/http handler interface
func (h *HealthcheckHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.ready {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
