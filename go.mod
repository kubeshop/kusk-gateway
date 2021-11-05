module github.com/kubeshop/kusk-gateway

go 1.16

require (
	github.com/envoyproxy/go-control-plane v0.9.10-0.20210907150352-cf90f659a021
	github.com/fsnotify/fsnotify v1.4.9
	github.com/getkin/kin-openapi v0.76.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/golang/protobuf v1.5.2
	github.com/jinzhu/copier v0.3.2
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/stretchr/testify v1.7.0
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
	k8s.io/api v0.22.3
	k8s.io/apimachinery v0.22.3
	k8s.io/client-go v0.22.3
	sigs.k8s.io/controller-runtime v0.10.0
)
