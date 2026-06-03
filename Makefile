API_DOCS ?= $(CURDIR)/bin/gen-crd-api-reference-docs
BRANCH ?= $(shell git tag --sort=-version:refname | grep -v '^manager/' | grep -v '^helm/' | head -1)

.PHONY: gen-api-docs
gen-api-docs: $(API_DOCS)
	@TEMPDIR=$$(mktemp -d) && \
	trap 'rm -rf "$$TEMPDIR"' EXIT && \
	echo "Extracting Go types from $(BRANCH)..." && \
	git archive $(BRANCH) -- api/ go.mod go.sum | tar -x -C "$$TEMPDIR" && \
	if [ ! -f "$$TEMPDIR/api/v1alpha1/doc.go" ]; then \
	  git show main:api/v1alpha1/doc.go > "$$TEMPDIR/api/v1alpha1/doc.go"; \
	fi && \
	echo "Downloading dependencies..." && \
	cd "$$TEMPDIR" && \
	go mod download && \
	echo "Generating API reference..." && \
	$(API_DOCS) \
	  -config $(CURDIR)/hack/api-reference/config.json \
	  -template-dir $(CURDIR)/hack/api-reference/template \
	  -api-dir ./api/v1alpha1 \
	  -out-file $(CURDIR)/includes/api-reference.html && \
	echo "Done: $(CURDIR)/includes/api-reference.html"

$(API_DOCS):
	GOBIN=$(CURDIR)/bin go install github.com/ahmetb/gen-crd-api-reference-docs@latest
