include Makefile.variables

$(smoketests):
	go mod tidy
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)

check-basic_auth:
	kubectl apply -f ../examples/ext-authz
	kubectl wait deployment --namespace=default ext-authz-http-service --for condition=Available=True --timeout=3m
	@echo "sleeping for 32 seconds ..."
	@sleep 32s
	kubectl get pods -A
	@echo
	-go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	@echo
	kubectl get pods -A
	@echo
	-kubectl logs $(shell kubectl get pods --namespace default | grep default | awk '{print $$2}')
	@echo
	-kubectl logs --namespace kusk-gateway-pr-483 --all-containers=true $(shell kubectl get pods --namespace kusk-gateway-pr-483 | grep kusk-gateway-manager | awk '{print $$2}')
	@echo
	kubectl delete -f ../examples/ext-authz

check-cache:
	kubectl apply -f ./samples/cache/cache-upstream.yaml
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	kubectl delete -f ./samples/cache/cache-upstream.yaml

sandbox:
	@docker build samples/hello-world/hello-world-container/ -t localhost:50000/hello-world:smoke
	@docker push localhost:50000/hello-world:smoke
	kubectl apply -f samples/hello-world/deployment.yaml
