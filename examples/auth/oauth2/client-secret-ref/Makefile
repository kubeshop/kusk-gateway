.PHONY: all
all:
	-kubectl delete -f ./secret.yaml
	-kubectl delete -f ./manifests.yaml
	-kubectl delete -f ./api.yaml
	-kubectl apply -f ./secret.yaml
	-kubectl apply -f ./manifests.yaml
	-kubectl apply -f ./api.yaml
