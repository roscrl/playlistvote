# [playlistvote.com](https://playlistvote.com)

vote on user submitted Spotify playlists

![config/readme/showoff.png](config/readme/showoff.png)

## Setup

Refer to the [Makefile](Makefile) for setup details

Set up the environment by installing Zig for the built-in C cross-compiler to Linux (or any C compiler for the 'make build' target) and Node.js for browser testing.

`make tools && make hotreload-mock`

## Dependencies

### Frontend

Server side rendered Go templates with `html/template`

`tailwindcss` styling

`hotwire/turbo` and `alpinejs` for frontend JS, both vendored

### Production

`go-sqlite3` as database driver (requires `zig cc` to compile x86 from ARM)

`sqlc` for generating Go code from [SQL queries](db/query.sql)

`prominentcolor` for extracting prominent colors from playlist images

`newrelic/go-agent` for application monitoring

### Development

`is` for assertions

`fsnotify` for watching Go template changes in dev mode without recompiling

#### Browser Tests

`node/npm` (requires [.node](browsertests/.node-version))

`make test-browser`

`playwright` for browser automation

## Deploy

The application is deploy on a [VPS](https://specbranch.com/posts/one-big-server/)

#### VPS Setup

- Set `VPS_IP` variable in the Makefile
- Run `make new-vps`

### Cloudflare

- Set SSL Full
- Add an A record in the DNS settings pointing to VPS IP

## Miscellaneous

#### Structure Inspiration

[Mat Ryer - How I write HTTP services after eight years talk](https://www.youtube.com/watch?v=XGVZ0Ip4XPM)  
[Mat Ryer - Deep dive of real application](https://www.youtube.com/watch?v=VRZZeJwIAIM)  
[Mat Ryer - Twitter thread](https://twitter.com/matryer/status/1445013230858952705?lang=en-GB)
