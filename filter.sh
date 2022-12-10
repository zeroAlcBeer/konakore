
GOBUILD=CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags '-w -s'  -o bin/filter cmd/filter/main.go
export wpath="C:\Users\cheer\Pictures\Saved Pictures"
export dsn="root:please_change@tcp(192.168.0.110:3307)/konakore?charset=utf8mb4&parseTime=True&loc=Local"
bin/filter