embed:
	cd chat && bun run build
	rm -rf lib/httpapi/chat && mkdir -p lib/httpapi/chat && touch lib/httpapi/chat/marker
	cp -r chat/out/. lib/httpapi/chat/

build: embed
	go build -o agentapi main.go
