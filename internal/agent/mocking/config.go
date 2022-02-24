package mocking

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/clbanning/mxj"
	"github.com/getkin/kin-openapi/openapi3"
)

var (
	JsonMediaTypePattern = regexp.MustCompile("^application/.*json$")
	XmlMediaTypePattern  = regexp.MustCompile("^application/.*xml$")
	TextMediaTypePattern = regexp.MustCompile("^text/.*$")
)

// Mapping of Mock ID to MockResponses
type MockConfig map[string]*MockResponse

func NewMockConfig() *MockConfig {
	return &MockConfig{}
}

func (m MockConfig) GetMockResponse(mockID string) *MockResponse {
	return m[mockID]
}

func (m MockConfig) AddMockResponse(mockID string, resp *MockResponse) error {
	if _, ok := m[mockID]; ok {
		return fmt.Errorf("mock response with ID %s already exists", mockID)
	}
	m[mockID] = resp
	return nil
}

func (m MockConfig) GenerateMockResponse(op *openapi3.Operation) (*MockResponse, error) {

	// https://swagger.io/docs/specification/describing-responses/
	// We iterate over each response until found only ONE candidate for the mocking:
	// if there is the single success (2xx) code without the schema and example - use it to return that simple code in the mocked response.
	// if there is success code with the "example" field - use this to create the mocked response body.
	// if there is success code with the "examples" field - use the first element to create the mocked response body.
	// otherwise if none found - fail, this operation must be excluded from the mocking specifically.
	mockResp := NewMockResponse()
	for respCode, respRef := range op.Responses {
		// We don't handle non 2xx codes, skip if found
		if !strings.HasPrefix(respCode, "2") {
			continue
		}
		// Note that we don't handle wildcards, e.g. '2xx' - this is allowed in OpenAPI, but we need the exact status code.
		statusCode, err := strconv.Atoi(respCode)
		if err != nil {
			return nil, fmt.Errorf("cannot convert the response code %s to int: %w", respCode, err)
		}
		mockResp.StatusCode = statusCode

		// The first found http code is a mock if it doesn't have any response body (e.g. just return 201)
		if respRef.Value.Content == nil {
			return mockResp, nil
		}
		// Otherwise we search for the example in each media type.
		// https://swagger.io/docs/specification/media-types/
		for mediaType, mediaTypeValue := range respRef.Value.Content {
			var exampleContent interface{}
			switch {
			case mediaTypeValue.Example != nil:
				exampleContent = mediaTypeValue.Example
			case mediaTypeValue.Examples != nil:
				// Get only the first returned example.
				// Note that this is not the stable order, sort it first if needed.
				for _, value := range mediaTypeValue.Examples {
					exampleContent = value
					break
				}
			default:
				// no example nor examples are present, skip this
				continue
			}
			if exampleContent != nil {
				examplebytes, err := marshallExampleContent(mediaType, exampleContent)
				if err != nil {
					return nil, fmt.Errorf("failure marshaling example content: %w", err)
				}
				mockResp.MediaTypeData[mediaType] = examplebytes
			}
		}
	}
	// Empty examples - don't set mock
	if len(mockResp.MediaTypeData) == 0 {
		return nil, fmt.Errorf("neither the body example nor a simple success (e.g. 200) code is present for mocking generation")
	}
	return mockResp, nil
}

func marshallExampleContent(format string, exampleContent interface{}) ([]byte, error) {
	switch {
	case JsonMediaTypePattern.MatchString(format):
		return json.Marshal(exampleContent)
	case XmlMediaTypePattern.MatchString(format):
		if object, isObject := exampleContent.(map[string]interface{}); isObject {
			xml := mxj.Map(object)
			return xml.Xml()
		} else {
			return mxj.AnyXml(exampleContent, "root")
		}
	case TextMediaTypePattern.MatchString(format):
		if bytes, ok := exampleContent.([]byte); ok {
			return bytes, nil
		}
		if s, ok := exampleContent.(string); ok {
			return []byte(s), nil
		}
		// If it can't just be converted to string explicitly, call String method for that type if present.
		if s, ok := exampleContent.(fmt.Stringer); ok {
			return []byte(s.String()), nil
		}
		return nil, fmt.Errorf("cannot serialise %s into string", format)

	}
	return nil, fmt.Errorf("unsupported format type %s", format)
}
