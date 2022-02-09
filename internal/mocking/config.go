package mocking

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type MockConfig struct {
	mockResponses map[string]*MockResponse
}

func NewMockConfig() *MockConfig {
	return &MockConfig{
		mockResponses: make(map[string]*MockResponse),
	}
}

func (m *MockConfig) GetMockResponse(mockID string) *MockResponse {
	return m.mockResponses[mockID]
}

func (m *MockConfig) SetMockResponse(mockID string, resp *MockResponse) error {
	if _, ok := m.mockResponses[mockID]; ok {
		return fmt.Errorf("mock response with ID %s already exists", mockID)
	}
	m.mockResponses[mockID] = resp
	return nil
}

func (m *MockConfig) GenerateMockResponse(op *openapi3.Operation) (*MockResponse, error) {

	// https://swagger.io/docs/specification/describing-responses/
	// We iterate over each response until found only ONE candidate for the mocking:
	// * if there is the single success (2xx) code without the schema and example - use it to return that simple code in the mocked response.
	// * if there is success code with the example - use this to create the mocked response body.
	mockResp := NewMockResponse()
	for respCode, respRef := range op.Responses {
		// We don't handle non 2xx codes, skip if found
		if !strings.HasPrefix(respCode, "2") {
			continue
		}
		// Note that we don't handle wildcards, e.g. '2xx' - this is allowed in OpenAPI, but we need the exact status code.
		statusCode, err := strconv.Atoi(respCode)
		if err != nil {
			return nil, err
		}
		mockResp.StatusCode = statusCode

		// The first found http code is a mock if it doesn't have any response body (e.g. just return 201)
		if respRef.Value.Content == nil {
			return mockResp, nil
		}
		// Otherwise we search for the example in each media type.
		// https://swagger.io/docs/specification/media-types/
		for mediaType, mediaTypeValue := range respRef.Value.Content {

			if mediaTypeValue.Example == nil {
				continue
			}
			switch mediaType {
			// TODO: case for other json variants
			case "application/json":
				examplebytes, err := json.Marshal(mediaTypeValue.Example)
				if err != nil {
					return nil, fmt.Errorf("failure marshalling example content: %w", err)
				}
				mockResp.MediaTypeData[mediaType] = examplebytes
			//TODO
			case "application/xml":
			case "text/html":
			case "text/plain":
			default:
				return nil, fmt.Errorf("unsupported media type %s", mediaType)
			}
		}
	}
	// Empty examples - don't set mock
	if len(mockResp.MediaTypeData) == 0 {
		return nil, fmt.Errorf("neither the body example nor a simple success (e.g. 200) code is present for mocking generation")
	}
	return mockResp, nil
}
