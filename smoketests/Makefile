include Makefile.variables

$(smoketests): 
	go mod tidy
	go test -count=1 -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)