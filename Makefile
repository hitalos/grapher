build:
	go generate ./public
	CGO_ENABLED=0 go build -ldflags '-s -w' -o dist/ .

dev:
	go generate -tags dev ./public
	ENV=dev go run -tags dev .

clean:
	rm -rf dist/

lint:
	go vet ./...
	golangci-lint run ./...

sec:
	govulncheck ./...
	gosec ./...
	grype .
	trivy fs .

install_deps:
	npm ci
	go mod tidy
