apc:
	go run cmd/cli/main.go 

test:
	go run cmd/cli/main.go -a graphql --url ./test --config ./cmd/cli 