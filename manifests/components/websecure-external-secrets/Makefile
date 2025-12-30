# WebSecure External Secrets Namespace Component Makefile
# This component only creates the namespace - External Secrets Operator will be deployed via ArgoCD

NAMESPACE ?= websecure-external-secrets
OUTPUT_DIR = ../../../talos-manifests

# Output file
FINAL_MANIFEST = $(OUTPUT_DIR)/websecure-external-secrets-namespace.yaml

.PHONY: all build clean

all: build

# Build namespace manifest
build:
	@mkdir -p $(OUTPUT_DIR)
	@echo "🏗️  Building WebSecure External Secrets namespace manifest..."
	@echo "---" > $(FINAL_MANIFEST)
	@echo "apiVersion: v1" >> $(FINAL_MANIFEST)
	@echo "kind: Namespace" >> $(FINAL_MANIFEST)
	@echo "metadata:" >> $(FINAL_MANIFEST)
	@echo "  name: $(NAMESPACE)" >> $(FINAL_MANIFEST)
	@echo "  labels:" >> $(FINAL_MANIFEST)
	@echo "    name: $(NAMESPACE)" >> $(FINAL_MANIFEST)
	@echo "    app.kubernetes.io/name: external-secrets" >> $(FINAL_MANIFEST)
	@echo "    app.kubernetes.io/component: namespace" >> $(FINAL_MANIFEST)
	@echo "✅ Namespace manifest created: $(FINAL_MANIFEST)"

clean:
	@echo "🧹 Cleaning WebSecure External Secrets namespace component..."
	@rm -f $(FINAL_MANIFEST)
