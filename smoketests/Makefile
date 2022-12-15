include Makefile.variables

$(smoketests):
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)

check-weighted_cluster:
	kubectl apply -f ./weighted_cluster/deployments.yaml
	kubectl wait deployment --namespace default traffic-splitting-httpbin --for condition=Available=True --timeout=3m
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	kubectl delete -f ./weighted_cluster/deployments.yaml

check-auth_jwt:
	kubectl apply -f ../examples/auth/jwt/oauth0/manifests.yaml
	kubectl wait deployment --namespace default auth-jwt-oauth0-go-httpbin --for condition=Available=True --timeout=3m
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	kubectl delete -f ../examples/auth/jwt/oauth0/manifests.yaml

check-cache:
	kubectl apply -f ./samples/cache/cache-upstream.yaml
	@echo "sleeping for 2s ..."
	@sleep 2
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	kubectl delete -f ./samples/cache/cache-upstream.yaml

check-basic_auth:
	-kubectl delete -f ./basic_auth/basic_auth_api.yaml
	kubectl apply -f ../examples/ext-authz/http-basic-auth/
	kubectl wait --namespace default deployment ext-authz-http-basic-auth --for condition=Available=True --timeout=3m
	@echo "sleeping for 6s ..."
	@sleep 6
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	kubectl delete -f ../examples/ext-authz/http-basic-auth/

check-openapi-path:
	kubectl apply -f ./openapi-path/openapi-path-with-auth-deployments.yaml
	kubectl wait deployment ext-authz-http-basic-auth-openapi-path-with-auth --namespace=default --for condition=Available=True --timeout=3m
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	kubectl delete -f ./openapi-path/openapi-path-with-auth-deployments.yaml

check-auth_oauth2:
	kubectl apply -f ../examples/auth/oauth2/authorization-code-grant/manifests.yaml
	kubectl wait deployment --namespace default auth-oauth2-oauth0-authorization-code-grant-go-httpbin --for condition=Available=True --timeout=3m
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	kubectl delete -f ../examples/auth/oauth2/authorization-code-grant/manifests.yaml
