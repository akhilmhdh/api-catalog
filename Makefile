apc:
	go run cmd/cli/main.go run

test-json:
	go run cmd/cli/main.go run -a openapi --url ./test/petstore-v3.json --config ./test

test-v2-openapi:
	go run cmd/cli/main.go run -a openapi --url https://petstore.swagger.io/v2/swagger.json --config ./test
