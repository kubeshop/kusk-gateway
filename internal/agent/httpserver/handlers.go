// package server provides the Agent HTTP server, which is the service, configured with the Agent Management Service.
package httpserver

import (
	"net/http"
	"sync"

	"github.com/kubeshop/kusk-gateway/internal/agent/mocking"
)

type mainHandler struct {
	mockConfig *mocking.MockConfig
	mu         *sync.RWMutex
}

func NewMainHandler() *mainHandler {
	return &mainHandler{mockConfig: mocking.NewMockConfig(), mu: &sync.RWMutex{}}
}

func (h *mainHandler) SetMockConfig(mockConfig *mocking.MockConfig) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.mockConfig = mockConfig
}

func (h *mainHandler) GetMockResponse(mockID string) *mocking.MockResponse {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.mockConfig.GetMockResponse(mockID)
}

// ServerHTTP implements the standard net/http handler interface
func (h *mainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mockID := r.Header.Get(HeaderMockID)
	// Fail if no mockID header in the request
	if mockID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	mockResponse := h.GetMockResponse(mockID)
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
