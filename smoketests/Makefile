# .PHONY: $(smoketests)
include Makefile.variables

$(smoketests):
	@echo $@
	go test -v github.com/kubeshop/kusk-gateway/smoketests/$(subst check-,,$@)