run:
	go run . --config ./config/.dev

run-mock:
	go run . --config ./config/.dev.mock

hotreload:
	air -c ./config/.air.toml & make tailwind-watch

hotreload-mock:
	air -c ./config/.air.mock.toml & make tailwind-watch

tailwind-watch:
	./bin/tailwindcss -i ./views/assets/main.css -o ./views/assets/dist/main.css --watch --config ./config/tailwind.config.js

generate:
	./bin/tailwindcss -i ./views/assets/main.css -o ./views/assets/dist/main.css --config ./config/tailwind.config.js
	./bin/esbuild views/assets/dist/js/alpine-3.12.0/alpine.js --minify --outfile=views/assets/dist/js/alpine-3.12.0/alpine.min.js
	./bin/esbuild views/assets/dist/js/alpine-3.12.0/intersect.js --minify --outfile=views/assets/dist/js/alpine-3.12.0/intersect.min.js
	./bin/esbuild views/assets/dist/js/turbo-7.3.0/dist/turbo.es2017-esm.js --minify --outfile=views/assets/dist/js/turbo-7.3.0/dist/turbo.es2017-esm.min.js
	cd ./db && sqlc generate

lint:
	golangci-lint run .

format:
	gofumpt -l -w .

test:
	go test -v ./...

test-browser:
	cd browsertests/ && npm run test

build: generate lint format test
	go build -o bin/app .

build-amd64: generate lint format test
	CC="zig cc -target x86_64-linux-musl" CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bin/app .

build-arm64: generate lint format test
	CC="zig cc -target aarch64-linux-musl" CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o bin/app .

build-quick: test
	go build -o bin/app .

VPS_IP=5.161.84.223
SERVICE_NAME=playlist-vote.service

make ssh:
	ssh root@$(VPS_IP)

new-vps:
	# Run VPS Setup Script in README.md
	make caddy-service-reload
	make db-copy-over
	make app-service-reload
	make deploy

caddy-service-reload:
	scp -r ./config/caddy.service root@$(VPS_IP):/lib/systemd/system/caddy.service
	ssh root@$(VPS_IP) "systemctl daemon-reload"
	ssh root@$(VPS_IP) "systemctl restart caddy"

caddy-reload:
	scp -r ./config/Caddyfile root@$(VPS_IP):/etc/caddy/Caddyfile
	ssh root@$(VPS_IP) "systemctl reload caddy"

db-copy-over:
	rsync -avz --ignore-existing ./db/playlist-vote.db root@$(VPS_IP):~/db/

app-service-reload:
	scp -r ./config/$(SERVICE_NAME) root@$(VPS_IP):/lib/systemd/system/$(SERVICE_NAME)
	ssh root@$(VPS_IP) "systemctl daemon-reload"
	ssh root@$(VPS_IP) "systemctl restart $(SERVICE_NAME)"

upload: build-amd64
	ssh root@$(VPS_IP) "mkdir -p ~/new"
	scp -r bin/app root@$(VPS_IP):~/new/app

deploy: upload
	ssh root@$(VPS_IP) "mkdir -p ~/archive"
	ssh root@$(VPS_IP) "if [ -d ~/app ]; then mv ~/app ~/archive/app_$$(date +"%Y%m%d%H%M%S"); fi"
	ssh root@$(VPS_IP) "mv ~/new/app ~/app"
	ssh root@$(VPS_IP) "systemctl restart $(SERVICE_NAME)"

logs-prod:
	ssh root@$(VPS_IP) "journalctl -u $(SERVICE_NAME) -f"

############################################################################################################

# Separately install Zig for its built in C cross compiler to linux (or any c compiler for make target build-arm64)
# Separately install NodeJS for browsertests/ browser testing
make tools:
	go install github.com/kyleconroy/sqlc/cmd/sqlc@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/cosmtrek/air@latest
	make tooling-esbuild
	make tooling-tailwind

# MacOS ARM specific
tooling-tailwind:
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v3.3.1/tailwindcss-macos-arm64
	chmod +x tailwindcss-macos-arm64
	mv tailwindcss-macos-arm64 tailwindcss
	mv tailwindcss ./bin

tooling-esbuild:
	curl -fsSL https://esbuild.github.io/dl/latest | sh
	mv esbuild ./bin
