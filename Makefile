SHELL=/bin/bash -e -o pipefail
PWD = $(shell pwd)

COOKIECUTTER_TEST_OUTPUT = my_awesome_project

.PHONY: pulumi-vars k9s kubeconfig talosconfig talosctl kubectl out download tidy fmt lint test test-cookiecutter govulncheck clean manifests manifests-clean manifests-help help

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

kubectl: kubeconfig ## Run kubectl for the current cluster
	@$(KUBECTL) $(filter-out $@,$(MAKECMDGOALS))

out:
	@mkdir -p out/build

download: ## Downloads the dependencies
	@go mod download

tidy: ## Cleans up go.mod and go.sum
	@go mod tidy

fmt: ## Formats all code with go fmt
	@go fmt ./...

lint: fmt $(GOLANGCI_LINT) download ## Lints all code with golangci-lint
	@go tool golangci-lint run

test: ## Runs all tests
	@go test $(ARGS) ./...

test-cookiecutter: ## Test cookiecutter template by generating a project and running make lint
	@rm -rf $(COOKIECUTTER_TEST_OUTPUT) && \
	cookiecutter . --no-input skip_go_mod_tidy=True && \
	cd $(COOKIECUTTER_TEST_OUTPUT) && \
	go mod edit -replace github.com/exivity/pulumi-hcloud-k8s=../ && \
    go mod tidy && \
	make lint && \
	rm -rf $(COOKIECUTTER_TEST_OUTPUT)

govulncheck: ## Vulnerability detection using govulncheck
	@go run golang.org/x/vuln/cmd/govulncheck ./...

clean: ## Cleans up everything
	@rm -rf bin out

manifests: manifests-clean ## Build all Talos extra manifests
	@$(MAKE) -C manifests build

manifests-clean: ## Clean all generated manifests
	@$(MAKE) -C manifests clean

manifests-help: ## Show manifest build help
	@$(MAKE) -C manifests help

help: ## Shows the help
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
        awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''
