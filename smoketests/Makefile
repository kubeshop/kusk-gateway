include Makefile.variables

$(smoketests):
	go mod tidy
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)

check-cache:
	kubectl apply -f ./samples/cache/cache-upstream.yaml
	@echo "sleeping for 2s ..."
	@sleep 2
	kubectl get svc -A
	-go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
 	-minikube --profile kgw ip && ([ $$? -eq 0 ] && curl -v "http://$(shell minikube --profile kgw ip)/uuid") || echo 'failed to get `/uuid`'
	sleep 1
	-minikube --profile kgw ip && ([ $$? -eq 0 ] && curl -v "http://$(shell minikube --profile kgw ip)/uuid") || echo 'failed to get `/uuid`'
	sleep 1
	-minikube --profile kgw ip && ([ $$? -eq 0 ] && curl -v "http://$(shell minikube --profile kgw ip)/uuid") || echo 'failed to get `/uuid`'
	sleep 1
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
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	kubectl delete -f ../examples/auth/oauth2/authorization-code-grant/manifests.yaml

sandbox:
	@docker build samples/hello-world/hello-world-container/ -t localhost:50000/hello-world:smoke
	@docker push localhost:50000/hello-world:smoke
	kubectl apply -f samples/hello-world/deployment.yaml
