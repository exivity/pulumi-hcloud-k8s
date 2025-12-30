# Cilium CNI Component Makefile

NAMESPACE ?= kube-system
RELEASE_NAME ?= cilium
BUILD_DIR = ../../build/cilium
OUTPUT_DIR = ../../../talos-manifests

# Helm repository details
HELM_REPO_NAME = cilium
HELM_REPO_URL = https://helm.cilium.io/
HELM_CHART = cilium/cilium

# Build directories for different variants
BUILD_DIR_KUBE_PROXY = $(BUILD_DIR)/kube-proxy
BUILD_DIR_KUBE_PROXY_GATEWAY = $(BUILD_DIR)/kube-proxy-gateway

# Output files
FINAL_MANIFEST_KUBE_PROXY = $(OUTPUT_DIR)/cilium-with-kube-proxy.yaml
FINAL_MANIFEST_KUBE_PROXY_GATEWAY = $(OUTPUT_DIR)/cilium-with-kube-proxy-gateway.yaml

.PHONY: all helm-repo-add helm-repo-update template-kube-proxy template-kube-proxy-gateway build build-kube-proxy build-kube-proxy-gateway clean

all: build

# Build all variants
build: build-kube-proxy build-kube-proxy-gateway

# Add Helm repository
helm-repo-add:
	@echo "Adding Cilium Helm repository..."
	helm repo add $(HELM_REPO_NAME) $(HELM_REPO_URL) || true

# Update Helm repository
helm-repo-update: helm-repo-add
	@echo "Updating Helm repositories..."
	helm repo update $(HELM_REPO_NAME)

# Template the Helm chart with kube-proxy (without GatewayAPI)
template-kube-proxy: helm-repo-update
	@mkdir -p $(BUILD_DIR_KUBE_PROXY)
	@echo "Templating Cilium with kube-proxy (without GatewayAPI)..."
	helm template $(RELEASE_NAME) $(HELM_CHART) \
		--namespace $(NAMESPACE) \
		--create-namespace \
		--set ipam.mode=kubernetes \
		--set kubeProxyReplacement=false \
		--set securityContext.capabilities.ciliumAgent="{CHOWN,KILL,NET_ADMIN,NET_RAW,IPC_LOCK,SYS_ADMIN,SYS_RESOURCE,DAC_OVERRIDE,FOWNER,SETGID,SETUID}" \
		--set securityContext.capabilities.cleanCiliumState="{NET_ADMIN,SYS_ADMIN,SYS_RESOURCE}" \
		--set cgroup.autoMount.enabled=false \
		--set cgroup.hostRoot=/sys/fs/cgroup \
		--output-dir $(BUILD_DIR_KUBE_PROXY)

# Template the Helm chart with kube-proxy and GatewayAPI
template-kube-proxy-gateway: helm-repo-update
	@mkdir -p $(BUILD_DIR_KUBE_PROXY_GATEWAY)
	@echo "Templating Cilium with kube-proxy and GatewayAPI..."
	helm template $(RELEASE_NAME) $(HELM_CHART) \
		--namespace $(NAMESPACE) \
		--create-namespace \
		--set ipam.mode=kubernetes \
		--set kubeProxyReplacement=false \
		--set securityContext.capabilities.ciliumAgent="{CHOWN,KILL,NET_ADMIN,NET_RAW,IPC_LOCK,SYS_ADMIN,SYS_RESOURCE,DAC_OVERRIDE,FOWNER,SETGID,SETUID}" \
		--set securityContext.capabilities.cleanCiliumState="{NET_ADMIN,SYS_ADMIN,SYS_RESOURCE}" \
		--set cgroup.autoMount.enabled=false \
		--set cgroup.hostRoot=/sys/fs/cgroup \
		--set gatewayAPI.enabled=true \
		--set gatewayAPI.enableAlpn=true \
		--set gatewayAPI.enableAppProtocol=true \
		--output-dir $(BUILD_DIR_KUBE_PROXY_GATEWAY)

# Build manifest with kube-proxy (without GatewayAPI)
build-kube-proxy: template-kube-proxy
	@mkdir -p $(OUTPUT_DIR)
	@echo "Building Cilium with kube-proxy manifest..."
	@echo "---" > $(FINAL_MANIFEST_KUBE_PROXY)
	@echo "apiVersion: v1" >> $(FINAL_MANIFEST_KUBE_PROXY)
	@echo "kind: Namespace" >> $(FINAL_MANIFEST_KUBE_PROXY)
	@echo "metadata:" >> $(FINAL_MANIFEST_KUBE_PROXY)
	@echo "  name: $(NAMESPACE)" >> $(FINAL_MANIFEST_KUBE_PROXY)
	@echo "---" >> $(FINAL_MANIFEST_KUBE_PROXY)
	@find $(BUILD_DIR_KUBE_PROXY) -name "*.yaml" -exec cat {} \; >> $(FINAL_MANIFEST_KUBE_PROXY)

# Build manifest with kube-proxy and GatewayAPI
build-kube-proxy-gateway: template-kube-proxy-gateway
	@mkdir -p $(OUTPUT_DIR)
	@echo "Building Cilium with kube-proxy and GatewayAPI manifest..."
	@echo "---" > $(FINAL_MANIFEST_KUBE_PROXY_GATEWAY)
	@echo "apiVersion: v1" >> $(FINAL_MANIFEST_KUBE_PROXY_GATEWAY)
	@echo "kind: Namespace" >> $(FINAL_MANIFEST_KUBE_PROXY_GATEWAY)
	@echo "metadata:" >> $(FINAL_MANIFEST_KUBE_PROXY_GATEWAY)
	@echo "  name: $(NAMESPACE)" >> $(FINAL_MANIFEST_KUBE_PROXY_GATEWAY)
	@echo "---" >> $(FINAL_MANIFEST_KUBE_PROXY_GATEWAY)
	@find $(BUILD_DIR_KUBE_PROXY_GATEWAY) -name "*.yaml" -exec cat {} \; >> $(FINAL_MANIFEST_KUBE_PROXY_GATEWAY)

# Install with kube-proxy via Helm (for reference, not used in manifest generation)
install-kube-proxy: helm-repo-update
	@echo "Installing Cilium with kube-proxy via Helm..."
	helm upgrade --install $(RELEASE_NAME)-kube-proxy $(HELM_CHART) \
		--namespace $(NAMESPACE) \
		--create-namespace \
		--set ipam.mode=kubernetes \
		--set kubeProxyReplacement=false \
		--set securityContext.capabilities.ciliumAgent="{CHOWN,KILL,NET_ADMIN,NET_RAW,IPC_LOCK,SYS_ADMIN,SYS_RESOURCE,DAC_OVERRIDE,FOWNER,SETGID,SETUID}" \
		--set securityContext.capabilities.cleanCiliumState="{NET_ADMIN,SYS_ADMIN,SYS_RESOURCE}" \
		--set cgroup.autoMount.enabled=false \
		--set cgroup.hostRoot=/sys/fs/cgroup

# Install with kube-proxy and GatewayAPI via Helm (for reference, not used in manifest generation)
install-kube-proxy-gateway: helm-repo-update
	@echo "Installing Cilium with kube-proxy and GatewayAPI via Helm..."
	helm upgrade --install $(RELEASE_NAME)-kube-proxy-gateway $(HELM_CHART) \
		--namespace $(NAMESPACE) \
		--create-namespace \
		--set ipam.mode=kubernetes \
		--set kubeProxyReplacement=false \
		--set securityContext.capabilities.ciliumAgent="{CHOWN,KILL,NET_ADMIN,NET_RAW,IPC_LOCK,SYS_ADMIN,SYS_RESOURCE,DAC_OVERRIDE,FOWNER,SETGID,SETUID}" \
		--set securityContext.capabilities.cleanCiliumState="{NET_ADMIN,SYS_ADMIN,SYS_RESOURCE}" \
		--set cgroup.autoMount.enabled=false \
		--set cgroup.hostRoot=/sys/fs/cgroup \
		--set gatewayAPI.enabled=true \
		--set gatewayAPI.enableAlpn=true \
		--set gatewayAPI.enableAppProtocol=true

# Legacy install target (defaults to kube-proxy without GatewayAPI)
install: install-kube-proxy

clean:
	rm -rf $(BUILD_DIR)
