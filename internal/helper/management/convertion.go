package management

import "github.com/kubeshop/kusk-gateway/internal/helper/mocking"

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
