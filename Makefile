default:
	@go build -o commitlore main.go

test:
	@go test ./...
testv:
	@go test -v ./...

