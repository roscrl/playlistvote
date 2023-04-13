clean:
	go clean

lint:
	golangci-lint run .

format:
	gofumpt -l -w .

build: clean lint format
	make build-tailwind && CGO_ENABLED=1 go build -o bin/app .

build-amd64: clean lint format
	make build-tailwind && CC="zig cc -target x86_64-linux-musl" CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bin/app .

run: 
	go run .

test:
	go test -v ./...

hotreload:
	air -c ./config/.air.toml & make watch-tailwind

mock-hotreload:
	air -c ./config/.air.mock.toml & make watch-tailwind

generate:
	cd ./db && sqlc generate

build-tailwind:
	./bin/tailwindcss -i ./views/assets/main.css -o ./views/dist/main.css --config ./config/tailwind.config.js

watch-tailwind:
	./bin/tailwindcss -i ./views/assets/main.css -o ./views/assets/dist/main.css --watch --config ./config/tailwind.config.js

tooling-tailwind-mac-arm:
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v3.3.1/tailwindcss-macos-arm64
	chmod +x tailwindcss-macos-arm64
	mv tailwindcss-macos-arm64 tailwindcss
	mv tailwindcss ./bin

make tooling-dev: tooling-tailwind-mac-arm
	go install github.com/kyleconroy/sqlc/cmd/sqlc@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/cosmtrek/air@latest
