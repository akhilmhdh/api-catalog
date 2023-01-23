apc:
	go run cmd/cli/main.go run

test-json:
	go run cmd/cli/main.go run -a openapi --url ./test/petstore-v3.json --config ./test