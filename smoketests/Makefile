include Makefile.variables


$(smoketests): sandbox
	go mod tidy
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)
	kubectl delete -f samples/hello-world/deployment.yaml


sandbox: 
	@docker build samples/hello-world/hello-world-container/ -t localhost:50000/hello-world:smoke
	@docker push localhost:50000/hello-world:smoke
	kubectl apply -f samples/hello-world/deployment.yaml
	
