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
	// If no media type data (example) found - this is the simple http code in the response, write it and return
	if len(mockResponse.MediaTypeData) == 0 {
		w.WriteHeader(mockResponse.StatusCode)
		return
	}
	mediaTypes := getMediaTypes(mockResponse.MediaTypeData)
	defaultMediaType := getDefaultMediaType(mediaTypes)
	// Get media type from the request Accept header parsing and matching to the existing media content type.
	// If missing or not matched - use the first entry in the media content map.
	chosenMediaType := NegotiateContentType(r, mediaTypes, defaultMediaType)
	data := mockResponse.MediaTypeData[chosenMediaType]
	w.Header().Set("Content-Type", chosenMediaType)
	w.WriteHeader(mockResponse.StatusCode)
	w.Write(data)
}

func getMediaTypes(mediaTypesData map[string][]byte) []string {
	mediaTypes := make([]string, 0, len(mediaTypesData))
	for contentType := range mediaTypesData {
		mediaTypes = append(mediaTypes, contentType)
	}

	return mediaTypes
}

func getDefaultMediaType(mediaTypes []string) string {
	if len(mediaTypes) == 1 {
		return mediaTypes[0]
	}
	// Return any json-based mediaType as default
	for _, mediaType := range mediaTypes {
		if mocking.JsonMediaTypePattern.Match([]byte(mediaType)) {
			return mediaType
		}
	}
	// Otherwise return the first found
	return mediaTypes[0]
}
