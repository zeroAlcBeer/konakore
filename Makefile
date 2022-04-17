GOBUILD=CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags '-w -s'  -o

server:
	$(GOBUILD) bin/server cmd/server/main.go
syncer:
	$(GOBUILD) bin/syncer cmd/syncer/main.go