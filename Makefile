#########################
##  Local Development  ##
#########################

run:
	go run . --config ./config/.dev

run-mock:
	make reset-mock-db
	go run . --config ./config/.dev.mock

hotreload:
	air -c ./config/.air.toml & make tailwind-watch

hotreload-mock:
	make reset-mock-db
	air -c ./config/.air.mock.toml & make tailwind-watch

reset-mock-db:
	rm -f ./db/playlistvote-mock.db && rm -f ./db/playlistvote-mock.db-shm && rm -f ./db/playlistvote-mock.db-wal
	cp ./db/playlistvote-mock-corpus.db ./db/playlistvote-mock.db

tailwind-watch:
	./bin/tailwindcss -i ./views/assets/main.css -o ./views/assets/dist/main.css --watch --config ./config/tailwind.config.js

generate:
	./bin/tailwindcss -i ./views/assets/main.css -o ./views/assets/dist/main.css --config ./config/tailwind.config.js
	./bin/esbuild views/assets/dist/js/vendor/stimulus-3.2.1/stimulus.js --minify --outfile=views/assets/dist/js/vendor/stimulus-3.2.1/stimulus.min.js
	./bin/esbuild views/assets/dist/js/vendor/turbo-7.3.0/dist/turbo.es2017-esm.js --minify --outfile=views/assets/dist/js/vendor/turbo-7.3.0/dist/turbo.es2017-esm.min.js
	cd ./db && sqlc generate

lint:
	golangci-lint run .

format:
	gofumpt -l -w .

test:
	go test -v ./...

#########################
#####    Builds     #####
#########################

build: generate lint format test
	go build -o bin/app .

build-amd64: generate lint format test
	CC="zig cc -target x86_64-linux-musl" CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bin/app .

build-arm64: generate lint format test
	CC="zig cc -target aarch64-linux-musl" CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o bin/app .

build-quick: test
	go build -o bin/app .

#########################
##### Remote server #####
#########################

VPS_IP=5.161.84.223
APP_FOLDER=~/playlistvote
SERVICE_NAME=playlistvote.service
LOCAL_SQLITE_DB_PATH=./db/playlistvote.db
USER=root

ssh:
	ssh $(USER)@$(VPS_IP)

vps-new:
	ssh $(USER)@$(VPS_IP) "mkdir -p $(APP_FOLDER)"
	make vps-dependencies
	make caddy-reload
	make caddy-service-reload
	make db-copy-over
	make app-service-reload
	make deploy

vps-dependencies:
	ssh $(USER)@$(VPS_IP) "sudo apt-get update && sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https && curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg && curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list && sudo apt update && sudo apt install caddy"

caddy-service-reload:
	scp -r ./config/caddy.service $(USER)@$(VPS_IP):/lib/systemd/system/caddy.service
	ssh $(USER)@$(VPS_IP) "systemctl daemon-reload"
	ssh $(USER)@$(VPS_IP) "systemctl restart caddy"

caddy-reload:
	scp -r ./config/Caddyfile $(USER)@$(VPS_IP):/etc/caddy/Caddyfile
	ssh $(USER)@$(VPS_IP) "systemctl reload caddy"

db-copy-over:
	rsync -avz --ignore-existing $(LOCAL_SQLITE_DB_PATH) $(USER)@$(VPS_IP):$(APP_FOLDER)/db/

db-copy-over-force:
	rsync -avz $(LOCAL_SQLITE_DB_PATH) $(USER)@$(VPS_IP):$(APP_FOLDER)/db/

db-copy-prod:
	rsync -avz --ignore-existing $(USER)@$(VPS_IP):$(APP_FOLDER)/db/ $(LOCAL_SQLITE_DB_PATH).prod

app-service-reload:
	scp -r ./config/$(SERVICE_NAME) $(USER)@$(VPS_IP):/lib/systemd/system/$(SERVICE_NAME)
	ssh $(USER)@$(VPS_IP) "systemctl daemon-reload"
	ssh $(USER)@$(VPS_IP) "systemctl restart $(SERVICE_NAME)"

upload: build-amd64
	ssh $(USER)@$(VPS_IP) "mkdir -p $(APP_FOLDER)/new"
	scp -r bin/app $(USER)@$(VPS_IP):$(APP_FOLDER)/new/app

deploy: upload
	ssh $(USER)@$(VPS_IP) "mkdir -p $(APP_FOLDER)/archive"
	ssh $(USER)@$(VPS_IP) "if [ -d $(APP_FOLDER)/app ]; then mv $(APP_FOLDER)/app $(APP_FOLDER)/archive/app_$$(date +"%Y%m%d%H%M%S"); fi"
	ssh $(USER)@$(VPS_IP) "mv $(APP_FOLDER)/new/app $(APP_FOLDER)/app"
	ssh $(USER)@$(VPS_IP) "systemctl restart $(SERVICE_NAME)"

logs-prod:
	ssh $(USER)@$(VPS_IP) "journalctl -u $(SERVICE_NAME) -f"

logs-caddy-prod:
	ssh $(USER)@$(VPS_IP) "journalctl -u caddy -f"

#########################
#####    Tooling    #####
#########################

tools:
	go install github.com/kyleconroy/sqlc/cmd/sqlc@v1.17.2
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.2
	go install mvdan.cc/gofumpt@v0.5.0
	go install github.com/cosmtrek/air@v1.43.0
	mkdir -p ./bin/
	make tooling-esbuild
	make tooling-tailwind
	echo "Remember to install Zig for the built-in C cross-compiler to Linux (or any C compiler for the 'make build' target)"

tooling-tailwind:
	# MacOS ARM specific
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v3.3.1/tailwindcss-macos-arm64
	chmod +x tailwindcss-macos-arm64
	mv tailwindcss-macos-arm64 tailwindcss
	mv tailwindcss ./bin/

tooling-esbuild:
	curl -fsSL https://esbuild.github.io/dl/v0.17.17 | sh
	mv esbuild ./bin/
