version := 0.1.0

apc:
	go run "-X main.version=${version}" cmd/cli/main.go run

test-json:
	go run cmd/cli/main.go run -a openapi --schema ./test/petstore-v3.json --config ./test

test-v2-openapi:
	go run cmd/cli/main.go run -a openapi --schema https://petstore.swagger.io/v2/swagger.json --config ./test

generate-plugin-zip:
	go run -ldflags "-X main.version=${version}" scripts/plugin_zipper/main.go  

build:
	go build -ldflags "-X main.version=${version}"  -o ./apic cmd/cli/main.go 
