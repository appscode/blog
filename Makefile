.PHONY: run
run:
	@yqq w -i config.dev.yaml params.search_api_key --tag '!!str' $(GOOGLE_CUSTOM_SEARCH_API_KEY)
	hugo server --config=config.dev.yaml --buildDrafts --buildFuture

.PHONY: assets
assets: hugo-tools
	$(HUGO_TOOLS) docs-aggregator --only-assets
	find ./data -name "*.json" -exec sed -i 's/https:\/\/cdn.appscode.com\/images/\/assets\/images/g' {} \;

.PHONY: fmt
fmt:
	hugo-tools fmt-frontmatter ./content

.PHONY: verify
verify: fmt
	hugo-tools tag-stats ./content --invalid-only
	@if !(git diff --exit-code HEAD); then \
		echo "files are out of date, run make fmt"; exit 1; \
	fi

.PHONY: gen-draft
gen-draft:
	rm -rf public
	@yqq w -i config.yaml params.search_api_key --tag '!!str' $(GOOGLE_CUSTOM_SEARCH_API_KEY)
	hugo --config=config.yaml --buildDrafts --buildFuture
	@yqq w -i config.yaml params.search_api_key --tag '!!str' '_replace_'

.PHONY: qa
qa: gen-draft
	firebase use default
	firebase deploy

.PHONY: gen-prod
gen-prod:
	rm -rf public
	@yqq w -i config.yaml params.search_api_key --tag '!!str' $(GOOGLE_CUSTOM_SEARCH_API_KEY)
	hugo --minify --config=config.yaml
	@yqq w -i config.yaml params.search_api_key --tag '!!str' '_replace_'

.PHONY: release
release: gen-prod
	firebase use prod
	firebase deploy
	firebase use default

HUGO_TOOLS = $(shell pwd)/bin/hugo-tools
.PHONY: hugo-tools
hugo-tools: ## Download hugo-tools locally if necessary.
	$(call go-get-tool,$(HUGO_TOOLS),appscodelabs/hugo-tools,v0.2.24)

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
curl -fsSL -o $(1) https://github.com/$(2)/releases/download/$(3)/$${bin}; \
chmod +x $(1); \
}
endef
