package mocking_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/kubeshop/kusk-gateway/internal/agent/mocking"
	"github.com/kubeshop/kusk-gateway/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func loadOperation(t *testing.T, path string) *openapi3.Operation {
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	parser := spec.NewParser(nil)
	apiSpec, err := parser.ParseFromReader(f)
	if err != nil {
		t.Fatal(err)
	}

	opts, err := spec.GetOptions(apiSpec)
	if err != nil {
		t.Fatal(err)
	}
	opts.FillDefaults()
	if err := opts.Validate(); err != nil {
		t.Fatal(err)
	}
	op := apiSpec.Paths["/mocked/{id}"].Operations()["GET"]

	return op
}

func TestMockConfigSingleExample(t *testing.T) {
	m := mocking.NewMockConfig()

	op := loadOperation(t, "testdata/example.yaml")
	resp, err := m.GenerateMockResponse(op)
	assert.NoError(t, err)
	fmt.Printf("%#v", resp)

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"completed\":true,\"order\":13,\"title\":\"Mocked JSON title\",\"url\":\"http://mockedURL.com\"}", string(resp.MediaTypeData["application/json"]))
}

func TestMockConfigMultipleExamples(t *testing.T) {
	m := mocking.NewMockConfig()

	op := loadOperation(t, "testdata/examples.yaml")
	resp, err := m.GenerateMockResponse(op)
	assert.NoError(t, err)
	fmt.Printf("%#v", resp)

	assert.Equal(t, 200, resp.StatusCode)
	assert.NotEmpty(t, resp)
}
