.PHONY: recreate-api
recreate-api:
	-kubectl delete -f ./api.yaml
	-kubectl delete -f ./manifests.yaml
	-kubectl apply -f ./api.yaml
	-kubectl apply -f ./manifests.yaml
