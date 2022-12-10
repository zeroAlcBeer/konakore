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
nsfw-run:
	make nsfw
	export wpath="C:\Users\cheer\Pictures\Saved Pictures"
	export dsn="root:please_change@tcp(192.168.0.110:3307)/konakore?charset=utf8mb4&parseTime=True&loc=Local"
	bin/nsfw
