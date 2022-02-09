package mockserver

import (
	"net/http"

	"github.com/kubeshop/kusk-gateway/internal/mocking"
)

const (
	// Server Hostname and port are not configurable since the manager needs to know them for the Envoy cluster creation
	ServerHostname           string = "127.0.0.1"
	ServerPort               uint32 = 8090
	HeaderMockID                    = "X-Kusk-Mock-ID"
	HeaderMockResponseInsert        = "X-Kusk-Mocked"
)

type MockResponses struct {
	responses map[string]*mocking.MockResponse
}

func NewMockResponses() *MockResponses {
	return &MockResponses{responses: make(map[string]*mocking.MockResponse)}
}

func (m *MockResponses) GetResponse(mockID string) *mocking.MockResponse {
	return m.responses[mockID]
}

// HTTP Handler to pass to the mux
type MockHTTPHandler struct {
	mockResponses *MockResponses
}

func NewMockHTTPHandler() *MockHTTPHandler {
	return &MockHTTPHandler{mockResponses: NewMockResponses()}
}

// ServerHTTP implements the standard net/http handler interface
func (m *MockHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: fail if the header is missing
	mockID := r.Header.Get(HeaderMockID)
	// TODO: Fail if the response is missing
	mockResponse := m.mockResponses.GetResponse(mockID)

	// TODO: detect content type for the user using its request Accept Header
	mediaType := "application/json"
	// TODO: fail if missing media type
	data := mockResponse.MediaTypeData[mediaType]
	w.Header().Set("Content-Type", mediaType)
	w.Header().Set(HeaderMockResponseInsert, "true")
	w.WriteHeader(mockResponse.StatusCode)
	w.Write(data)
}
