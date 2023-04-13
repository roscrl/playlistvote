![config/readme/example.png](config/readme/example.png)

## Setup


`make tooling-dev` Mac ARM specific setup just needed for Tailwind executable

`make mock-hotreload` 

`make hotreload` to hit Spotify API. API key required

## Dependencies

### Frontend

`tailwindcss` at build time, Node not required

`hotwire/turbo` for frontend JS, vendored

`alpinejs` for frontend JS, vendored

### Production 

`go-sqlite3` for database driver

`sqlc` for SQL code generation

`prominentcolor` for image color extraction of playlist covers

`newrelic/go-agent` for monitoring


### Development 

`is` for assertions  

`fsnotify` for watching file template changes for dev mode

## Structure Inspiration

[Mat Ryer - How I write HTTP services after eight years talk](https://www.youtube.com/watch?v=XGVZ0Ip4XPM)  
[Mat Ryer - Deep dive of real application](https://www.youtube.com/watch?v=VRZZeJwIAIM)
[Mat Ryer - Twitter thread](https://twitter.com/matryer/status/1445013230858952705?lang=en-GB)
