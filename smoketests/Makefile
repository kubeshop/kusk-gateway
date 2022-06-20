include Makefile.variables


$(smoketests):
	go mod tidy
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)

check-basic_auth:
	kubectl apply -f ../examples/ext-authz
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	kubectl delete -f ../examples/ext-authz
	
sandbox: 
	@docker build samples/hello-world/hello-world-container/ -t localhost:50000/hello-world:1.0
	@docker push localhost:50000/hello-world:1.0
	kubectl apply -f samples/hello-world/deployment.yaml
	
