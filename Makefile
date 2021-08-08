GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags '-w -s'  -o
BIN=bin/konakore
SOURCE=cmd/gallery/main.go

BIN2=bin/syncer
SOURCE2=cmd/syncer/main.go

docker:
	$(GOBUILD) $(BIN) $(SOURCE)
syncer:
	$(GOBUILD) $(BIN2) $(SOURCE2)