# WPT with Polyfills

A reverse proxy to [wpt.live](https://github.com/web-platform-tests/wpt) that injects polyfills from [polyfill.io](https://polyfill.io/v3/)

## Run it locally

```
go get https://github.com/romainmenke/wpt-polyfilled
DEV=true PORT=10345 wpt-polyfilled
```

visit : [bs-local:10345](http://bs-local.com:10345)

## Deploy it somewhere

ENV Vars :

| key | value |
|-----|-------|
| PUBLIC_ADDR | <your-domain> |
| PORT | <your public http port> |

DNS :

| type | name | content |
|-----|-------|---------|
| CNAME | wpt | <your host> |
| CNAME | wpt-a | <your host> |
| CNAME | wpt-b | <your host> |

visit : `wpt.<your.domain>`

## How it works

URLs to wpt.live are rewritten so everything goes through the proxy.
The proxy injects a script tag for polyfill.io.


## Notice

This was created in a couple of hours. Not all tests where validated.
I expect issues where urls are encoded and thus not replaced correctly.
