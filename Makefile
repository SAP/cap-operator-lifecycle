LOCALBIN ?= $(shell pwd)/bin
API_DOCS ?= $(LOCALBIN)/gen-crd-api-reference-docs
API_DOCS_VERSION ?= latest
BRANCH ?= main

.PHONY: gen-api-docs
gen-api-docs: ## Generate API reference documentation from Go types. Override branch with BRANCH=<name>.
	bash hack/api-reference/generate.sh $(BRANCH)

.PHONY: gen-crd-api-reference-docs
gen-crd-api-reference-docs: $(API_DOCS) ## Download gen-crd-api-reference-docs locally if necessary.
$(API_DOCS): $(LOCALBIN)
	test -s $(LOCALBIN)/gen-crd-api-reference-docs || GOBIN=$(LOCALBIN) go install github.com/ahmetb/gen-crd-api-reference-docs@$(API_DOCS_VERSION)

$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
