.PHONY: run
run:
	@yq w -i config.dev.yaml params.search_api_key --tag '!!str' $(GOOGLE_CUSTOM_SEARCH_API_KEY)
	hugo server --config=config.dev.yaml --buildDrafts --buildFuture

.PHONY: docs
docs:
	hugo-tools docs-aggregator
	find ./data -name "*.json" -exec sed -i 's/https:\/\/cdn.appscode.com\/images/\/assets\/images/g' {} \;

.PHONY: gen
gen:
	rm -rf public
	@yq w -i config.dev.yaml params.search_api_key --tag '!!str' $(GOOGLE_CUSTOM_SEARCH_API_KEY)
	hugo --config=config.dev.yaml --buildDrafts --buildFuture
	@yq w -i config.dev.yaml params.search_api_key --tag '!!str' '_replace_'

.PHONY: qa
qa: gen
	firebase use default
	firebase deploy

.PHONY: gen-prod
gen-prod:
	rm -rf public
	@yq w -i config.yaml params.search_api_key --tag '!!str' $(GOOGLE_CUSTOM_SEARCH_API_KEY)
	hugo --minify --config=config.yaml
	@yq w -i config.yaml params.search_api_key --tag '!!str' '_replace_'

.PHONY: release
release: gen-prod
	firebase use prod
	firebase deploy
	firebase use default
