generate:
	oapi-codegen -config oapi-codegen.yaml openapi/openapi.yml

run:
	go run ./cmd/app

tidy:
	go mod tidy
