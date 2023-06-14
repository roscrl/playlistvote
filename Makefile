#########################
##  Local Development  ##
#########################

run-mock:
	go run . --config ./config/.dev.mock

hotreload-mock:
	air -c ./config/.air.mock.toml & make tailwind-watch

run:
	go run . --config ./config/.dev

hotreload:
	air -c ./config/.air.toml & make tailwind-watch

tailwind-watch:
	./bin/tailwindcss -i ./core/views/assets/main.css -o ./core/views/assets/dist/main.css --watch --config ./config/tailwind.config.js

generate:
	./bin/tailwindcss -i ./core/views/assets/main.css -o ./core/views/assets/dist/main.css --config ./config/tailwind.config.js
	./bin/esbuild core/views/assets/dist/js/vendor/stimulus-3.2.1/stimulus.js --minify --outfile=core/views/assets/dist/js/vendor/stimulus-3.2.1/stimulus.min.js
	./bin/esbuild core/views/assets/dist/js/vendor/turbo-7.3.0/dist/turbo.es2017-esm.js --minify --outfile=core/views/assets/dist/js/vendor/turbo-7.3.0/dist/turbo.es2017-esm.min.js
	cd ./core/db && sqlc generate

lint:
	golangci-lint run --config config/.golangci.yml

format:
	gofumpt -l -w . && gci write -s standard -s default ./..

.PHONY: test
test:
	go test -v ./...

test-browser-slow:
	go test -v ./... -rod=show,slow=1s,trace

bench:
	go test -run=^$ -bench=. ./...

pprof:
	go tool pprof -http=:8080 bin/profile.pprof

#########################
#####    Scripts    #####
#########################

generate-mock-playlists:
	go run scripts/generatemockplaylists.go

#########################
#####    Builds     #####
#########################

build: generate format lint test
	go build -o bin/app .

build-amd64: generate format lint test
	CC="zig cc -target x86_64-linux-musl" CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bin/app .

build-arm64: generate format lint test
	CC="zig cc -target aarch64-linux-musl" CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o bin/app .

build-quick:
	go build -o bin/app .

#########################
#####      VPS      #####
#########################

USER=root

APP_NAME=playlistvote
APP_FOLDER=~/$(APP_NAME)

APP_CADDY_PATH=$(APP_NAME).caddy
SERVICE_NAME=$(APP_NAME).service

DB_NAME=$(APP_NAME)
LOCAL_SQLITE_DB_PATH=./core/db/$(DB_NAME).db
LOCAL_SQLITE_SHM_DB_PATH=./core/db/$(DB_NAME).db-shm
LOCAL_SQLITE_WAL_DB_PATH=./core/db/$(DB_NAME).db-wal

CLOUDFLARE_ZONE_ID=3849d0e239cfff8040f0dceaf0071e4a

ssh:
	ssh $(USER)@$(VPS_IP)

vps-new:
	ssh $(USER)@$(VPS_IP) "mkdir -p $(APP_FOLDER)"
	make vps-dependencies
	make caddy-root-config
	make caddy-cert
	make caddy-service-reload
	make caddy-reload
	make db-copy-over
	make app-service-reload
	make deploy

vps-dependencies:
	ssh $(USER)@$(VPS_IP) "sudo apt-get update && sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https && curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg && curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list && sudo apt -y update && sudo apt -y install caddy && sudo apt -y install lnav"

caddy-root-config:
	scp -r ./config/Caddyfile $(USER)@$(VPS_IP):/etc/caddy/Caddyfile

caddy-service-reload:
	scp -r ./config/caddy.service $(USER)@$(VPS_IP):/lib/systemd/system/caddy.service
	ssh $(USER)@$(VPS_IP) "systemctl daemon-reload"
	ssh $(USER)@$(VPS_IP) "systemctl restart caddy"

caddy-cert:
	scp -r ./config/public.pem $(USER)@$(VPS_IP):/etc/ssl/certs/$(APP_NAME).pem
	ssh $(USER)@$(VPS_IP) "mkdir -p /etc/ssl/private"
	scp -r ./config/private.pem $(USER)@$(VPS_IP):/etc/ssl/private/$(APP_NAME).pem

caddy-reload:
	scp -r ./config/$(APP_CADDY_PATH) $(USER)@$(VPS_IP):/etc/caddy/$(APP_CADDY_PATH)
	ssh $(USER)@$(VPS_IP) "systemctl reload caddy"

db-copy-prod:
	rsync -avz --ignore-existing $(USER)@$(VPS_IP):$(APP_FOLDER)/db/ $(LOCAL_SQLITE_DB_PATH).prod

db-copy-over:
	rsync -avz --ignore-existing $(LOCAL_SQLITE_DB_PATH) $(USER)@$(VPS_IP):$(APP_FOLDER)/db/

db-copy-over-force:
	ssh $(USER)@$(VPS_IP) "mkdir -p $(APP_FOLDER)/db/archive"
	ssh $(USER)@$(VPS_IP) "if [ -f $(APP_FOLDER)/db/$(DB_NAME).db ];     then mv $(APP_FOLDER)/db/$(DB_NAME).db     $(APP_FOLDER)/db/archive/$(DB_NAME)_$$(date +"%Y%m%d%H%M%S").db;     fi"
	ssh $(USER)@$(VPS_IP) "if [ -f $(APP_FOLDER)/db/$(DB_NAME).db-shm ]; then mv $(APP_FOLDER)/db/$(DB_NAME).db-shm $(APP_FOLDER)/db/archive/$(DB_NAME)_$$(date +"%Y%m%d%H%M%S").db-shm; fi"
	ssh $(USER)@$(VPS_IP) "if [ -f $(APP_FOLDER)/db/$(DB_NAME).db-wal ]; then mv $(APP_FOLDER)/db/$(DB_NAME).db-wal $(APP_FOLDER)/db/archive/$(DB_NAME)_$$(date +"%Y%m%d%H%M%S").db-wal; fi"
	rsync -avz $(LOCAL_SQLITE_DB_PATH) $(USER)@$(VPS_IP):$(APP_FOLDER)/db/
	rsync -avz $(LOCAL_SQLITE_SHM_DB_PATH) $(USER)@$(VPS_IP):$(APP_FOLDER)/db/
	rsync -avz $(LOCAL_SQLITE_WAL_DB_PATH) $(USER)@$(VPS_IP):$(APP_FOLDER)/db/

app-service-reload:
	scp -r ./config/$(SERVICE_NAME) $(USER)@$(VPS_IP):/lib/systemd/system/$(SERVICE_NAME)
	ssh $(USER)@$(VPS_IP) "systemctl daemon-reload"
	ssh $(USER)@$(VPS_IP) "systemctl restart $(SERVICE_NAME)"

upload: build-amd64
	ssh $(USER)@$(VPS_IP) "mkdir -p $(APP_FOLDER)/new"
	scp -r bin/app $(USER)@$(VPS_IP):$(APP_FOLDER)/new/app

deploy: upload
	ssh $(USER)@$(VPS_IP) "mkdir -p $(APP_FOLDER)/archive"
	ssh $(USER)@$(VPS_IP) "if [ -f $(APP_FOLDER)/app ]; then mv $(APP_FOLDER)/app $(APP_FOLDER)/archive/app_$$(date +"%Y%m%d%H%M%S"); fi"
	ssh $(USER)@$(VPS_IP) "mv $(APP_FOLDER)/new/app $(APP_FOLDER)/app"
	ssh $(USER)@$(VPS_IP) "systemctl restart $(SERVICE_NAME)"
	make purge-cache-prod

purge-cache-prod:
	curl -X POST https://api.cloudflare.com/client/v4/zones/$(CLOUDFLARE_ZONE_ID)/purge_cache \
		-H "X-Auth-Email: $(CLOUDFLARE_EMAIL)" \
		-H "X-Auth-Key: $(CLOUDFLARE_KEY)" \
		-H "Content-Type: application/json" \
		--data '{"purge_everything":true}'

logs-prod:
	echo "make ssh then run 'journalctl -u $(SERVICE_NAME) | lnav'"

logs-prod-tail:
	ssh $(USER)@$(VPS_IP) "journalctl -u $(SERVICE_NAME) -f"

logs-caddy-prod:
	ssh $(USER)@$(VPS_IP) "journalctl -u caddy -f"

#########################
#####    Tooling    #####
#########################

tools:
	go install github.com/kyleconroy/sqlc/cmd/sqlc@v1.18.0
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.2
	go install mvdan.cc/gofumpt@v0.5.0
	go install github.com/daixiang0/gci@v0.10.1
	go install github.com/cosmtrek/air@v1.43.0
	go install github.com/playwright-community/playwright-go/cmd/playwright
	playwright install --with-deps
	mkdir -p ./bin/
	make tooling-esbuild
	make tooling-tailwind
	echo "Remember to install Zig for the built-in C cross-compiler to Linux (or any C compiler for the 'make build' targets)"

tooling-esbuild:
	curl -fsSL https://esbuild.github.io/dl/v0.17.17 | sh
	mv esbuild ./bin/

# MacOS ARM specific
tooling-tailwind:
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v3.3.2/tailwindcss-macos-arm64
	chmod +x tailwindcss-macos-arm64
	mv tailwindcss-macos-arm64 tailwindcss
	mv tailwindcss ./bin/
