.PHONY: get deps test modtidy gofmt vet modlist

GO111MODULE=on

all: mhs5200a

mhs5200a: main.go mhs5200a.go script.go
	go build -o bin/mhs5200a

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

