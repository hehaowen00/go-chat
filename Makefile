.PHONY: client
client:
	cd cmd/client && go run .

.PHONY: server
server:
	cd cmd/server && go run .

.PHONY: lookup
lookup:
	cd cmd/lookup && go run .
