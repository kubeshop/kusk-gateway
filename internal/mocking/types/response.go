package types

type MockResponse struct {
	// 200, 201
	StatusCode int
	// application/json -> []byte
	// application/xml -> []byte
	MediaTypeData map[string][]byte
}

func NewMockResponse() *MockResponse {
	return &MockResponse{
		MediaTypeData: make(map[string][]byte),
	}
}
