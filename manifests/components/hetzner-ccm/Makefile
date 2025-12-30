# Hetzner Cloud Controller Manager Component Makefile

NAMESPACE ?= kube-system
RELEASE_NAME ?= hetzner-ccm
BUILD_DIR = ../../build/hetzner-ccm
OUTPUT_DIR = ../../../talos-manifests

# Helm repository details
HELM_REPO_NAME = hcloud
HELM_REPO_URL = https://charts.hetzner.cloud
HELM_CHART = hcloud/hcloud-cloud-controller-manager

# Build directories for different variants
BUILD_DIR_PUBLIC = $(BUILD_DIR)/public
BUILD_DIR_PRIVATE = $(BUILD_DIR)/private

# Values files
VALUES_PUBLIC = values.yaml
VALUES_PRIVATE = values.yaml,values-private.yaml

# Output files
FINAL_MANIFEST_PUBLIC = $(OUTPUT_DIR)/hetzner-ccm-public.yaml
FINAL_MANIFEST_PRIVATE = $(OUTPUT_DIR)/hetzner-ccm-private.yaml

.PHONY: all helm-repo-add helm-repo-update template-public template-private build build-public build-private clean

all: build

# Build both variants
build: build-public build-private

# Add Helm repository
helm-repo-add:
	@echo "Adding Hetzner Cloud Helm repository..."
	helm repo add $(HELM_REPO_NAME) $(HELM_REPO_URL) || true

# Update Helm repository
helm-repo-update: helm-repo-add
	@echo "Updating Helm repositories..."
	helm repo update $(HELM_REPO_NAME)

# Template the Helm chart for public network (default configuration)
template-public: helm-repo-update
	@mkdir -p $(BUILD_DIR_PUBLIC)
	@echo "Templating Hetzner CCM Helm chart for public network..."
	helm template $(RELEASE_NAME) $(HELM_CHART) \
		--namespace $(NAMESPACE) \
		--create-namespace \
		--values $(VALUES_PUBLIC) \
		--output-dir $(BUILD_DIR_PUBLIC)

# Template the Helm chart for private network (with networking enabled)
template-private: helm-repo-update
	@mkdir -p $(BUILD_DIR_PRIVATE)
	@echo "Templating Hetzner CCM Helm chart for private network..."
	helm template $(RELEASE_NAME) $(HELM_CHART) \
		--namespace $(NAMESPACE) \
		--create-namespace \
		--values $(VALUES_PRIVATE) \
		--output-dir $(BUILD_DIR_PRIVATE)

# Build public network manifest
build-public: template-public
	@mkdir -p $(OUTPUT_DIR)
	@echo "Building Hetzner CCM public network manifest..."
	@echo "---" > $(FINAL_MANIFEST_PUBLIC)
	@echo "apiVersion: v1" >> $(FINAL_MANIFEST_PUBLIC)
	@echo "kind: Namespace" >> $(FINAL_MANIFEST_PUBLIC)
	@echo "metadata:" >> $(FINAL_MANIFEST_PUBLIC)
	@echo "  name: $(NAMESPACE)" >> $(FINAL_MANIFEST_PUBLIC)
	@echo "---" >> $(FINAL_MANIFEST_PUBLIC)
	@find $(BUILD_DIR_PUBLIC) -name "*.yaml" -exec cat {} \; >> $(FINAL_MANIFEST_PUBLIC)

# Build private network manifest
build-private: template-private
	@mkdir -p $(OUTPUT_DIR)
	@echo "Building Hetzner CCM private network manifest..."
	@echo "---" > $(FINAL_MANIFEST_PRIVATE)
	@echo "apiVersion: v1" >> $(FINAL_MANIFEST_PRIVATE)
	@echo "kind: Namespace" >> $(FINAL_MANIFEST_PRIVATE)
	@echo "metadata:" >> $(FINAL_MANIFEST_PRIVATE)
	@echo "  name: $(NAMESPACE)" >> $(FINAL_MANIFEST_PRIVATE)
	@echo "---" >> $(FINAL_MANIFEST_PRIVATE)
	@find $(BUILD_DIR_PRIVATE) -name "*.yaml" -exec cat {} \; >> $(FINAL_MANIFEST_PRIVATE)

# Install public network variant via Helm (for reference, not used in manifest generation)
install-public: helm-repo-update
	@echo "Installing Hetzner CCM (public network) via Helm..."
	helm upgrade --install $(RELEASE_NAME)-public $(HELM_CHART) \
		--namespace $(NAMESPACE) \
		--create-namespace \
		--values $(VALUES_PUBLIC)

# Install private network variant via Helm (for reference, not used in manifest generation)
install-private: helm-repo-update
	@echo "Installing Hetzner CCM (private network) via Helm..."
	helm upgrade --install $(RELEASE_NAME)-private $(HELM_CHART) \
		--namespace $(NAMESPACE) \
		--create-namespace \
		--values $(VALUES_PRIVATE)

# Legacy install target (defaults to public)
install: install-public

clean:
	rm -rf $(BUILD_DIR)
