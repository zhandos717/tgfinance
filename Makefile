OUTDIR  := bin
LDFLAGS := -ldflags="-s -w"
SERVER  ?= $(error SERVER is not set. Usage: make deploy SERVER=root@host)

.PHONY: build linux deploy test clean

## build — compile locally into bin/
build:
	mkdir -p $(OUTDIR)
	go build $(LDFLAGS) -o $(OUTDIR)/bot ./cmd/tgbot

## linux — cross-compile for Linux amd64 into bin/
linux:
	mkdir -p $(OUTDIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(OUTDIR)/bot-linux_amd64 ./cmd/tgbot

## deploy — build + upload + restart service
deploy: linux
	scp $(OUTDIR)/bot-linux_amd64 $(SERVER):/opt/aigate-bot
	ssh $(SERVER) "chmod +x /opt/aigate-bot && systemctl restart aigate-bot && systemctl is-active aigate-bot"

## test — run all tests
test:
	go test ./...

## clean — remove bin/
clean:
	rm -rf $(OUTDIR)
