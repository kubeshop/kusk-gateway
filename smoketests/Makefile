include Makefile.variables

$(smoketests):
	go mod tidy
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)

check-basic_auth:
	kubectl apply -f ../examples/ext-authz
	kubectl wait deployment --namespace=default ext-authz-http-service --for condition=Available=True --timeout=3m
	@echo "sleeping for 6s ..."
	@sleep 6
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	kubectl delete -f ../examples/ext-authz

check-cache:
	kubectl apply -f ./samples/cache/cache-upstream.yaml
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	kubectl delete -f ./samples/cache/cache-upstream.yaml

check-openapi-path:
	kubectl apply -f ./samples/openapi-path/openapi-path-with-auth-upstream.yaml
	kubectl wait deployment ext-authz-http-service-openapi-path-with-auth --namespace=default --for condition=Available=True --timeout=3m
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	kubectl delete -f ./samples/openapi-path/openapi-path-with-auth-upstream.yaml

sandbox:
	@docker build samples/hello-world/hello-world-container/ -t localhost:50000/hello-world:smoke
	@docker push localhost:50000/hello-world:smoke
	kubectl apply -f samples/hello-world/deployment.yaml
