# .PHONY: $(smoketests)
include Makefile.variables

test: 
	@echo ''

$(smoketests): test
	touch .timestamp
	# @echo github.com/kubeshop/kusk-gateway/smoketests/basic
	echo "TADA"
	go mod tidy
	go test -v github.com/kubeshop/kusk-gateway/smoketests/$@