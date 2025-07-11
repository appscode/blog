.PHONY: run
run:
	@yqq w -i config.dev.yaml params.search_api_key --tag '!!str' $(GOOGLE_CUSTOM_SEARCH_API_KEY)
	hugo server --config=config.dev.yaml --buildDrafts --buildFuture

.PHONY: assets
assets: hugo-tools
	$(HUGO_TOOLS) docs-aggregator --only-assets
	find ./data -name "*.json" -exec sed -i 's/https:\/\/cdn.appscode.com\/images/\/assets\/images/g' {} \;
	rm -rf static/files/cluster-api
	rm -rf static/files/cluster-api-provider-aws
	rm -rf static/files/cluster-api-provider-azure
	rm -rf static/files/cluster-api-provider-gcp
	rm -rf static/files/products/appscode/aws-marketplace
	rm -rf static/files/products/appscode/azure-marketplace
	rm -rf static/files/products/appscode/gcp-marketplace

.PHONY: fmt
fmt: hugo-tools
	$(HUGO_TOOLS) fmt-frontmatter ./content

.PHONY: tags
tags: hugo-tools
	$(HUGO_TOOLS) tag-stats ./content

.PHONY: verify
verify: fmt
	$(HUGO_TOOLS) tag-stats ./content --invalid-only
	@if !(git diff --exit-code HEAD); then \
		echo "files are out of date, run make fmt"; exit 1; \
	fi

.PHONY: gen-draft
gen-draft:
	rm -rf public
	hugo --config=config.dev.yaml --buildDrafts --buildFuture -d public/blog
	mv public/blog/404.html public/404.html

.PHONY: qa
qa: gen-draft
	firebase use default
	firebase deploy

.PHONY: gen-prod
gen-prod:
	rm -rf public
	hugo --minify --config=config.yaml -d public/blog
	mv public/blog/404.html public/404.html

.PHONY: release
release: gen-prod
	firebase use prod
	firebase deploy
	firebase use default

HUGO_TOOLS = $(shell pwd)/bin/hugo-tools
.PHONY: hugo-tools
hugo-tools: ## Download hugo-tools locally if necessary.
	$(call go-get-tool,$(HUGO_TOOLS),appscodelabs/hugo-tools)

# go-get-tool will 'curl' binary from GH repo $2 with version $3 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
OS=$$(echo `uname`|tr '[:upper:]' '[:lower:]'); \
ARCH=$$(uname -m); \
case $$ARCH in \
  armv5*) ARCH="armv5";; \
  armv6*) ARCH="armv6";; \
  armv7*) ARCH="arm";; \
  aarch64) ARCH="arm64";; \
  x86) ARCH="386";; \
  x86_64) ARCH="amd64";; \
  i686) ARCH="386";; \
  i386) ARCH="386";; \
esac; \
bin=hugo-tools-$${OS}-$${ARCH}; \
echo "Downloading $${bin}" ;\
mkdir -p $(PROJECT_DIR)/bin; \
curl -fsSL -o $(1) https://github.com/$(2)/releases/latest/download/$${bin}; \
chmod +x $(1); \
}
endef
