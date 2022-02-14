// package server provides the Helper HTTP server, which is the service, configured with the Helper Management Service.
package server

import (
	"net/http"

	"github.com/kubeshop/kusk-gateway/internal/helper/mocking"
)

const (
	// Server Hostname and port are not configurable since the manager needs to know them for the Envoy cluster creation
	ServerHostname           string = "127.0.0.1"
	ServerPort               uint32 = 8090
	HeaderMockID                    = "X-Kusk-Mock-ID"
	HeaderMockResponseInsert        = "X-Kusk-Mocked"
)

// HTTP Handler to pass to the mux
type HTTPHandler struct {
	mockConfig *mocking.MockConfig
}

func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{mockConfig: mocking.NewMockConfig()}
}

// ServerHTTP implements the standard net/http handler interface
func (m *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: fail if the header is missing
	mockID := r.Header.Get(HeaderMockID)
	// TODO: Fail if the response is missing
	mockResponse := m.mockConfig.GetMockResponse(mockID)

	// TODO: detect content type for the user using its request Accept Header
	mediaType := "application/json"
	// TODO: fail if missing media type
	data := mockResponse.MediaTypeData[mediaType]
	w.Header().Set("Content-Type", mediaType)
	w.Header().Set(HeaderMockResponseInsert, "true")
	w.WriteHeader(mockResponse.StatusCode)
	w.Write(data)
}
