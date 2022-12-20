package rate_limit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/smoketests/common"
	"github.com/kubeshop/kusk-gateway/smoketests/helpers"
)

const (
	testName      = "test-rate-limit"
	testNamespace = "default"
)

type RateLimitTestSuite struct {
	common.KuskTestSuite
	api *kuskv1.API
}

func (t *RateLimitTestSuite) SetupTest() {
	rawApi := common.ReadFile("./rate_limit.yaml")
	api := &kuskv1.API{}
	t.NoError(yaml.Unmarshal([]byte(rawApi), api))

	api.ObjectMeta.Name = testName
	api.ObjectMeta.Namespace = testNamespace
	api.Spec.Fleet.Name = helpers.APIFleetName
	api.Spec.Fleet.Namespace = helpers.APIFleetNamespace

	if err := t.Cli.Create(context.Background(), api, &client.CreateOptions{}); err != nil {
		if strings.Contains(err.Error(), `apis.gateway.kusk.io "test-rate-limit" already exists`) {
			// store `api` for deletion later
			t.api = api
			return
		}

		t.Fail(err.Error(), nil)
	}

	// store `api` for deletion later
	t.api = api

	// weird way to wait it out probably needs to be done dynamically
	t.T().Logf("Sleeping for %s", helpers.WaitBeforeStartingTest)
	time.Sleep(helpers.WaitBeforeStartingTest)
}

func (t *RateLimitTestSuite) TearDownSuite() {
	t.NoError(t.Cli.Delete(context.Background(), t.api, &client.DeleteOptions{}))
}

func (t *RateLimitTestSuite) TestRateLimitReached() {
	// We are expecting 429 Too Many Requests with a body of `local_rate_limited` once we do 3 requests.
	const (
		RateLimit = 2
	)

	loadBalancerIP := helpers.GetEnvoyFleetServiceLoadBalancerIP(&t.KuskTestSuite)
	url := fmt.Sprintf("http://%s/rate_limit", loadBalancerIP)

	// Do 2 requests then the next one will fail
	for x := 0; x < RateLimit; x++ {
		func() {
			req, err := http.NewRequest(http.MethodGet, url, nil)
			t.NoError(err)

			res, err := http.DefaultClient.Do(req)
			t.NoError(err)

			defer res.Body.Close()

			responseBody, err := io.ReadAll(res.Body)
			t.NoError(err)

			body := map[string]string{}
			t.NoError(json.Unmarshal(responseBody, &body))

			t.Equal(http.StatusOK, res.StatusCode)
			t.Equal(`rate-limited mocked response.`, body["message"])
		}()
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	t.NoError(err)

	res, err := http.DefaultClient.Do(req)
	t.NoError(err)

	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	t.NoError(err)

	t.Equal(http.StatusTooManyRequests, res.StatusCode)
	t.Equal("local_rate_limited", string(responseBody))
}

func TestRateLimitTestSuite(t *testing.T) {
	testSuite := RateLimitTestSuite{}
	suite.Run(t, &testSuite)
}
