/*
MIT License

Copyright (c) 2022 Kubeshop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package management

import "github.com/kubeshop/kusk-gateway/internal/agent/mocking"

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
