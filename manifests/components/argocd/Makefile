# ArgoCD Component Makefile

NAMESPACE ?= argocd
RELEASE_NAME ?= argocd
BUILD_DIR = ../../build/argocd
OUTPUT_DIR = ../../../talos-manifests

# Helm repository details
HELM_REPO_NAME = argo
HELM_REPO_URL = https://argoproj.github.io/argo-helm
HELM_CHART = argo/argo-cd

# Build directories for different variants
BUILD_DIR_STABLE = $(BUILD_DIR)/stable
BUILD_DIR_HA = $(BUILD_DIR)/ha

# Values files (can be created later for customization)
VALUES_STABLE = values-stable.yaml
VALUES_HA = values-ha.yaml

# Output files
FINAL_MANIFEST_STABLE = $(OUTPUT_DIR)/argocd-namespaced-stable.yaml
FINAL_MANIFEST_HA = $(OUTPUT_DIR)/argocd-namespaced-ha.yaml

.PHONY: all helm-repo-add helm-repo-update template-stable template-ha build build-stable build-ha clean

all: build

# Build both variants
build: build-stable build-ha

# Add Helm repository
helm-repo-add:
	@echo "Adding ArgoCD Helm repository..."
	helm repo add $(HELM_REPO_NAME) $(HELM_REPO_URL) || true

# Update Helm repository
helm-repo-update: helm-repo-add
	@echo "Updating Helm repositories..."
	helm repo update $(HELM_REPO_NAME)

# Template the Helm chart for stable deployment
template-stable: helm-repo-update
	@mkdir -p $(BUILD_DIR_STABLE)
	@echo "Templating ArgoCD Helm chart for stable deployment..."
	@if [ -f $(VALUES_STABLE) ]; then \
		helm template $(RELEASE_NAME) $(HELM_CHART) \
			--namespace $(NAMESPACE) \
			--create-namespace \
			--values $(VALUES_STABLE) \
			--output-dir $(BUILD_DIR_STABLE); \
	else \
		helm template $(RELEASE_NAME) $(HELM_CHART) \
			--namespace $(NAMESPACE) \
			--create-namespace \
			--output-dir $(BUILD_DIR_STABLE); \
	fi

# Template the Helm chart for HA deployment
template-ha: helm-repo-update
	@mkdir -p $(BUILD_DIR_HA)
	@echo "Templating ArgoCD Helm chart for HA deployment..."
	@if [ -f $(VALUES_HA) ]; then \
		helm template $(RELEASE_NAME) $(HELM_CHART) \
			--namespace $(NAMESPACE) \
			--create-namespace \
			--values $(VALUES_HA) \
			--output-dir $(BUILD_DIR_HA); \
	else \
		helm template $(RELEASE_NAME) $(HELM_CHART) \
			--namespace $(NAMESPACE) \
			--create-namespace \
			--set redis-ha.enabled=true \
			--set controller.replicas=1 \
			--set server.replicas=2 \
			--set repoServer.replicas=2 \
			--set applicationSet.replicas=2 \
			--output-dir $(BUILD_DIR_HA); \
	fi

# Build stable manifest
build-stable: template-stable
	@mkdir -p $(OUTPUT_DIR)
	@echo "Building ArgoCD stable manifest..."
	@echo "---" > $(FINAL_MANIFEST_STABLE)
	@echo "apiVersion: v1" >> $(FINAL_MANIFEST_STABLE)
	@echo "kind: Namespace" >> $(FINAL_MANIFEST_STABLE)
	@echo "metadata:" >> $(FINAL_MANIFEST_STABLE)
	@echo "  name: $(NAMESPACE)" >> $(FINAL_MANIFEST_STABLE)
	@echo "---" >> $(FINAL_MANIFEST_STABLE)
	@find $(BUILD_DIR_STABLE) -name "*.yaml" -exec cat {} \; >> $(FINAL_MANIFEST_STABLE)

# Build HA manifest
build-ha: template-ha
	@mkdir -p $(OUTPUT_DIR)
	@echo "Building ArgoCD HA manifest..."
	@echo "---" > $(FINAL_MANIFEST_HA)
	@echo "apiVersion: v1" >> $(FINAL_MANIFEST_HA)
	@echo "kind: Namespace" >> $(FINAL_MANIFEST_HA)
	@echo "metadata:" >> $(FINAL_MANIFEST_HA)
	@echo "  name: $(NAMESPACE)" >> $(FINAL_MANIFEST_HA)
	@echo "---" >> $(FINAL_MANIFEST_HA)
	@find $(BUILD_DIR_HA) -name "*.yaml" -exec cat {} \; >> $(FINAL_MANIFEST_HA)

# Install stable variant via Helm (for reference, not used in manifest generation)
install-stable: helm-repo-update
	@echo "Installing ArgoCD (stable) via Helm..."
	@if [ -f $(VALUES_STABLE) ]; then \
		helm upgrade --install $(RELEASE_NAME)-stable $(HELM_CHART) \
			--namespace $(NAMESPACE) \
			--create-namespace \
			--values $(VALUES_STABLE); \
	else \
		helm upgrade --install $(RELEASE_NAME)-stable $(HELM_CHART) \
			--namespace $(NAMESPACE) \
			--create-namespace; \
	fi

# Install HA variant via Helm (for reference, not used in manifest generation)
install-ha: helm-repo-update
	@echo "Installing ArgoCD (HA) via Helm..."
	@if [ -f $(VALUES_HA) ]; then \
		helm upgrade --install $(RELEASE_NAME)-ha $(HELM_CHART) \
			--namespace $(NAMESPACE) \
			--create-namespace \
			--values $(VALUES_HA); \
	else \
		helm upgrade --install $(RELEASE_NAME)-ha $(HELM_CHART) \
			--namespace $(NAMESPACE) \
			--create-namespace \
			--set redis-ha.enabled=true \
			--set controller.replicas=1 \
			--set server.replicas=2 \
			--set repoServer.replicas=2 \
			--set applicationSet.replicas=2; \
	fi

# Legacy install target (defaults to stable)
install: install-stable

clean:
	rm -rf $(BUILD_DIR)
