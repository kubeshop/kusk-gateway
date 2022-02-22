package management

import "github.com/kubeshop/kusk-gateway/internal/helper/mocking"

// MockConfigToProtoMockConfig is needed to do the convertion of mocking MockConfig type to protobuf generated one
func MockConfigToProtoMockConfig(mockConfig *mocking.MockConfig) *MockConfig {

	newMockConfig := &MockConfig{
		MockResponses: make(map[string]*MockResponse, len(*mockConfig)),
	}
	for mockID, mockResponse := range *mockConfig {
		// Create protobuf MockResponses with the status code
		newMockConfig.MockResponses[mockID] = &MockResponse{
			StatusCode:    uint32(mockResponse.StatusCode),
			MediaTypeData: make(map[string][]byte, len(mockResponse.MediaTypeData)),
		}
		// Fill its MediaTypeData mapping
		for mediaType, mediaTypeData := range mockResponse.MediaTypeData {
			newMockConfig.MockResponses[mockID].MediaTypeData[mediaType] = mediaTypeData
		}
	}
	return newMockConfig
}

// ProtoMockConfigToMockConfig is needed to do the convertion of protobuf MockConfig type to mocking MockConfig
func ProtoMockConfigToMockConfig(pbMockConfig *MockConfig) *mocking.MockConfig {

	newMockConfig := mocking.NewMockConfig()
	for mockID, mockResponse := range pbMockConfig.MockResponses {
		newResponse := mocking.NewMockResponse()
		newResponse.StatusCode = int(mockResponse.StatusCode)
		// Fill its MediaTypeData mapping
		for mediaType, mediaTypeData := range mockResponse.MediaTypeData {
			newResponse.MediaTypeData[mediaType] = mediaTypeData
		}
		newMockConfig.AddMockResponse(mockID, newResponse)
	}
	return newMockConfig
}
