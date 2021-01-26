.PHONY: get deps test modtidy gofmt vet modlist

GO111MODULE=on

all: mhs5200a

release: mhs5200a-linux-amd64 mhs5200a-win-amd64 mhs5200a-darwin-amd64

mhs5200a: main.go mhs5200a.go script.go
	go build -o bin/mhs5200a$(shell go env GOEXE)

mhs5200a-linux-amd64: main.go mhs5200a.go script.go
	env GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/mhs5200a

mhs5200a-win-amd64: main.go mhs5200a.go script.go
	env GOOS=windows GOARCH=amd64 go build -o bin/windows-amd64/mhs5200a.exe

mhs5200a-darwin-amd64: main.go mhs5200a.go script.go
	env GOOS=darwin GOARCH=amd64 go build -o bin/darwin-amd64/mhs5200a

get:
	go get -u $(PKG)

deps:
	go get -d ./...

test:
	go test 

modtidy:
	go mod tidy

gofmt:
	go fmt

vet:
	go vet

modlist:
	go list -m all

