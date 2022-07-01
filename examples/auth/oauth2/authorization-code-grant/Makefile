.PHONY: update-client_secret
update-client_secret:
	echo 'TODO(MBana): Update `client_secret` in the envoy config.'

.PHONY: recreate-api
recreate-api:
	-kubectl delete -f ./api.yaml
	-kubectl delete -f ./manifests.yaml
	-kubectl apply -f ./api.yaml
	-kubectl apply -f ./manifests.yaml
