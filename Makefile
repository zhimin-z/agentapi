CHAT_SOURCES_STAMP = chat/.sources.stamp
CHAT_SOURCES = $(shell find chat \( -path chat/node_modules -o -path chat/out -o -path chat/.next \) -prune -o -not -path chat/.sources.stamp -type f -print)

$(CHAT_SOURCES_STAMP): $(CHAT_SOURCES)
	@echo "Chat sources changed. Running build steps..."
	cd chat && bun run build
	rm -rf lib/httpapi/chat && mkdir -p lib/httpapi/chat && touch lib/httpapi/chat/marker
	cp -r chat/out/. lib/httpapi/chat/
	touch $@

.PHONY: embed
embed: $(CHAT_SOURCES_STAMP)
	@echo "Chat build is up to date."

.PHONY: build
build: embed
	go build -o agentapi main.go
