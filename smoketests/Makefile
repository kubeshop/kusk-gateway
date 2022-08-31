include Makefile.variables

$(smoketests):
	go mod tidy
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)

check-cache:
	kubectl apply -f ./samples/cache/cache-upstream.yaml
	@echo "sleeping for 2s ..."
	@sleep 2
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	kubectl delete -f ./samples/cache/cache-upstream.yaml

check-basic_auth:
	kubectl apply -f ../examples/ext-authz/http-basic-auth/
	kubectl wait --namespace default deployment ext-authz-http-basic-auth --for condition=Available=True --timeout=3m
	@echo "sleeping for 6s ..."
	@sleep 6
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	kubectl delete -f ../examples/ext-authz/http-basic-auth/

check-openapi-path:
	kubectl apply -f ./samples/openapi-path/openapi-path-with-auth-upstream.yaml
	kubectl wait deployment ext-authz-http-basic-auth-openapi-path-with-auth --namespace=default --for condition=Available=True --timeout=3m
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	kubectl delete -f ./samples/openapi-path/openapi-path-with-auth-upstream.yaml

check-auth_oauth2:
	kubectl apply -f ../examples/auth/oauth2/authorization-code-grant/manifests.yaml
	kubectl wait deployment --namespace default auth-oauth2-oauth0-authorization-code-grant-go-httpbin --for condition=Available=True --timeout=3m
	kubectl port-forward --namespace default deployments/default 19000:19000 & echo $$! > /tmp/kube-port-forward-check-auth_oauth2.pid
	sleep 2
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	kubectl delete -f ../examples/auth/oauth2/authorization-code-grant/manifests.yaml
	bash -c "trap 'pkill -F /tmp/kube-port-forward-check-auth_oauth2.pid' SIGINT SIGTERM ERR EXIT"

sandbox:
	@docker build samples/hello-world/hello-world-container/ -t localhost:50000/hello-world:smoke
	@docker push localhost:50000/hello-world:smoke
	kubectl apply -f samples/hello-world/deployment.yaml
