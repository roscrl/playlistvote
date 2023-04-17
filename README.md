![config/readme/example.png](config/readme/example.png)

## Setup

`make tools` 

`make hotreload-mock` 

`make hotreload`

## Dependencies

### Frontend

`tailwindcss` at build time, Node not required

`hotwire/turbo` for frontend JS, vendored

`alpinejs` for frontend JS, vendored

### Production 

`go-sqlite3` for database driver, requires CGO enabled and to compile to x86 from ARM requires `zig cc`

`sqlc` for SQL code generation

`prominentcolor` for image color extraction of playlist covers

`newrelic/go-agent` for monitoring


### Development 

`is` for assertions  

`fsnotify` for watching Go template changes for dev mode without recompiling

### Deploy

Single VPS server with Caddy, New Relic Agent, and a SQLite database.

#### VPS Setup Script (Debian)
```bash
# Caddy
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https &&
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg &&
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list &&
sudo apt update &&
sudo apt install caddy &&

# New Relic
curl -Ls https://download.newrelic.com/install/newrelic-cli/scripts/install.sh | bash && sudo NEW_RELIC_API_KEY=<KEY_HERE> NEW_RELIC_ACCOUNT_ID=<ACC_ID_HERE> /usr/local/bin/newrelic install &&

# SQLite
sudo apt install sqlite3
```

[Caddy Systemd Service](config/caddy.service) `caddy-service-reload`  
[Caddyfile](config/caddy/Caddyfile) `make caddy-reload`  

### Cloudflare

SSL Full  
DNS A Record set to VPS IP

## Code Structure Inspiration

[Mat Ryer - How I write HTTP services after eight years talk](https://www.youtube.com/watch?v=XGVZ0Ip4XPM)  
[Mat Ryer - Deep dive of real application](https://www.youtube.com/watch?v=VRZZeJwIAIM)  
[Mat Ryer - Twitter thread](https://twitter.com/matryer/status/1445013230858952705?lang=en-GB)
