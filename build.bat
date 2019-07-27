set GOOS=windows
go build -o bin/avatar-fight-server.exe cmd/all/main.go

set GOOS=linux
go build -o bin/avatar-fight-server cmd/all/main.go