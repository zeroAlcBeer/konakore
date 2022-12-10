GOBUILD=CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags '-w -s'  -o

server:
	go env -w GOPROXY="https://goproxy.cn,direct" && go env -w GOSUMDB=sum.golang.google.cn
	go mod tidy
	$(GOBUILD) bin/server cmd/server/main.go
syncer:env;
	go env -w GOPROXY="https://goproxy.cn,direct" && go env -w GOSUMDB=sum.golang.google.cn
	go mod tidy
	$(GOBUILD) bin/syncer cmd/syncer/main.go
nsfw:
	go env -w GOPROXY="https://goproxy.cn,direct" && go env -w GOSUMDB=sum.golang.google.cn
	go mod tidy
	$(GOBUILD) bin/nsfw cmd/nsfw/main.go
