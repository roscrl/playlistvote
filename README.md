# [playlistvote.com](https://playlistvote.com)

vote on user submitted Spotify playlists

![config/readme/showoff.png](config/readme/showoff.png)

## Setup

Refer to the [Makefile](Makefile) for setup details

Install Zig for the built-in C cross-compiler to Linux (or any C compiler for the 'make build' target)

`make tools && make hotreload-mock`

## Dependencies

### Frontend

Server side rendered Go templates with `html/template`

`tailwindcss` styling

`@hotwired/turbo` and `@hotwired/stimulus` for frontend JS, both vendored

### Production

`go-sqlite3` as database driver (requires `zig cc` to compile x86 from ARM)

`sqlc` for generating Go code from [SQL queries](core/db/query.sql)

`prominentcolor` for extracting prominent colors from playlist images

`newrelic/go-agent` for application monitoring

### Development

`is` for assertions

`fsnotify` for watching Go template changes in dev mode without recompiling

`rod` for browser tests

## Deploy

This application is deployed on a [VPS](https://specbranch.com/posts/one-big-server/)

#### VPS Setup

- Ensure `config/private.pem` exists (cloudflare origin certificate private key)
- Ensure `config/.prod` exists (app config)
- Set `VPS_IP` environment variable
- Set `CLOUDFLARE_EMAIL` environment variable
- Set `CLOUDFLARE_KEY` environment variable
- Run `make vps-new`

### Cloudflare

- Set SSL `Full (strict)`
- Add an A record in the DNS settings pointing to VPS IP
- Create Origin Certificate and place in `config/public.pem` & `config/private.pem`
- Enable Rate Limiting
  - `(http.request.uri.path contains "/")` 50 requests per 10s
- Enable [Bot Fight Mode](https://developers.cloudflare.com/bots/get-started/free/)
- Enable Page Rules Caching to respect `Cache-Control` headers returned
    - playlistvote.com/* Cache Level: Cache Everything
- Always Use HTTPS, Enable Brotli

### Hetzner

- Set firewall to allow only [Cloudflare IPs](https://www.cloudflare.com/en-gb/ips/) on port 443
- Set firewall to allow only personal IP on port 22

## Miscellaneous

#### Structure Inspiration

[Mat Ryer - How I write HTTP services after eight years talk](https://www.youtube.com/watch?v=XGVZ0Ip4XPM)  
[Mat Ryer - Deep dive of real application](https://www.youtube.com/watch?v=VRZZeJwIAIM)  
[Mat Ryer - Twitter thread](https://twitter.com/matryer/status/1445013230858952705?lang=en-GB)
