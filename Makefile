tidy:
	go mod tidy

codegen-api:
	oapi-codegen \
	-generate gin,strict-server,types,spec \
	-package http -o internal/api/http/http.gen.go docs/oapi/api.yaml

codegen-mocks:
	mockery

tests:
	go test -v -count=1 ./...

run:
	docker compose -f docker-compose.yaml up --build