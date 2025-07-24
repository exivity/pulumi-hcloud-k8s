SHELL=/bin/bash -e -o pipefail
PWD = $(shell pwd)

# Pulumi related variables
pulumi-vars:
	$(eval STACK_NAME=$(shell pulumi stack --show-name))
	$(eval KUBECONFIG=./$(STACK_NAME).kubeconfig.yml)
	$(eval KUBECONFIG_FLAG=--kubeconfig $(KUBECONFIG))
	$(eval KUBECTL=kubectl $(KUBECONFIG_FLAG))
	$(eval K9S=k9s $(KUBECONFIG_FLAG))
	$(eval HELM=helm $(KUBECONFIG_FLAG))
	$(eval TALOSCONFIG=./$(STACK_NAME).talosconfig.json)
	$(eval TALOSCTL=talosctl --talosconfig $(TALOSCONFIG))
	@echo "Pulumi variables set."

k9s: kubeconfig ## Run k9s for the current cluster
	@$(K9S) $(filter-out $@,$(MAKECMDGOALS))
%:
	@:

kubeconfig: pulumi-vars ## Get the kubeconfig for the current cluster
	@pulumi stack output kubeconfig --show-secrets > $(KUBECONFIG)
	@chmod 600 $(KUBECONFIG)

talosconfig: pulumi-vars ## Get the Talos config for the current cluster
	@pulumi stack output talosconfig --show-secrets > $(TALOSCONFIG)
	@chmod 600 $(TALOSCONFIG)

talosctl: talosconfig ## Run talosctl for the current cluster
	@$(TALOSCTL) $(filter-out $@,$(MAKECMDGOALS))

download: ## Downloads the dependencies
	@go mod download

fmt: ## Formats all code with go fmt
	@go fmt ./...

lint: fmt $(GOLANGCI_LINT) download ## Lints all code with golangci-lint
	@go tool -modfile=golangci-lint.mod golangci-lint run

test: ## Runs all tests
	@go test $(ARGS) ./...

clean: ## Cleans up everything
	@rm -rf bin

help: ## Shows the help
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
        awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''
