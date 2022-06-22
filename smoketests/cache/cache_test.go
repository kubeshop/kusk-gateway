package Cache

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"io"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/smoketests/common"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"

	"strings"
	"testing"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testName         = "test-cache"
	defaultName      = "default"
	defaultNamespace = "default"
)

type CacheTestSuite struct {
	common.KuskTestSuite
	api *kuskv1.API
}

func (t *CacheTestSuite) SetupTest() {
	rawApi := common.ReadFile("../samples/cache/cache.yaml")
	api := &kuskv1.API{}
	t.NoError(yaml.Unmarshal([]byte(rawApi), api))

	api.ObjectMeta.Name = testName
	api.ObjectMeta.Namespace = defaultNamespace
	api.Spec.Fleet.Name = defaultName
	api.Spec.Fleet.Namespace = defaultNamespace

	if err := t.Cli.Create(context.Background(), api, &client.CreateOptions{}); err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("apis.gateway.kusk.io %q already exists", testName)) {
			return
		}

		t.Fail(err.Error())
	}

	t.api = api // store `api` for deletion later

	duration := 4 * time.Second
	t.T().Logf("Sleeping for %s", duration)
	time.Sleep(duration) // weird way to wait it out probably needs to be done dynamically
}

func (t *CacheTestSuite) TestCacheCacheOn() {
	// We are expecting `cache-control: max-age=2` header and the value of `age` to increase over time.
	// the `uuid` will be cached up until `NoOfRequests`.
	const (
		NoOfRequests = 3
	)

	envoyFleetSvc := getEnvoyFleetSvc(t)
	url := fmt.Sprintf("http://%s/uuid", envoyFleetSvc.Status.LoadBalancer.Ingress[0].IP)

	uuidCached := ""
	// Do n requests that will get cached, i.e., they'll return the same uuid.
	for x := 0; x < NoOfRequests; x++ {
		func() {
			req, err := http.NewRequest("GET", url, nil)

			t.NoError(err)

			res, err := http.DefaultClient.Do(req)
			t.NoError(err)
			t.Equal(http.StatusOK, res.StatusCode)

			defer res.Body.Close()

			responseBody, err := io.ReadAll(res.Body)
			t.NoError(err)

			body := map[string]string{}
			t.NoError(json.Unmarshal(responseBody, &body))

			if body["uuid"] == "" {
				t.Fail("uuid is empty - expecting a uuid")
			}

			if uuidCached == "" && x == 0 {
				uuidCached = body["uuid"]
			}

			if uuidCached != body["uuid"] {
				t.Fail("uuid has changed - expecting the same uuid")
			}

			time.Sleep(1 * time.Second)
		}()
	}

	req, err := http.NewRequest("GET", url, nil)
	t.NoError(err)

	res, err := http.DefaultClient.Do(req)
	t.NoError(err)
	t.Equal(http.StatusOK, res.StatusCode)

	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	t.NoError(err)

	body := map[string]string{}
	t.NoError(json.Unmarshal(responseBody, &body))

	t.NotEqual(body["uuid"], uuidCached)
}

func (t *CacheTestSuite) TearDownSuite() {
	t.NoError(t.Cli.Delete(context.Background(), t.api, &client.DeleteOptions{}))
}

func TestCacheTestSuite(t *testing.T) {
	testSuite := CacheTestSuite{}
	suite.Run(t, &testSuite)
}

func getEnvoyFleetSvc(t *CacheTestSuite) *corev1.Service {
	t.T().Helper()

	envoyFleetSvc := &corev1.Service{}
	t.NoError(
		t.Cli.Get(context.Background(), client.ObjectKey{Name: defaultName, Namespace: defaultNamespace}, envoyFleetSvc),
	)

	return envoyFleetSvc
}
