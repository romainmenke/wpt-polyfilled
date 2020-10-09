# WPT with Polyfills

A reverse proxy for [wpt.live](https://github.com/web-platform-tests/wpt) that injects polyfills from [polyfill.io](https://polyfill.io/v3/)


## Hosted

[wpt.mysterious-mountain.stream](http://wpt.mysterious-mountain.stream/)


## Example

Compare these in Edge <=18 which does not support EventSource

[WPT Live   : EventSource URL](http://wpt.live/eventsource/eventsource-url.htm)

[Polyfilled : EventSource URL](http://wpt.mysterious-mountain.stream/eventsource/eventsource-url.htm)


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
| PUBLIC_ADDR | `<your-domain>` |
| PORT | `<your public http port>` |

DNS :

| type | name | content |
|-----|-------|---------|
| CNAME | wpt | `<your host>` |
| CNAME | wpt-a | `<your host>` |
| CNAME | wpt-b | `<your host>` |

visit : `wpt.<your.domain>`


## How it works

URLs to wpt.live are rewritten so everything goes through the proxy.
The proxy injects a script tag for polyfill.io.


## Notices

1. This was created in a couple of hours. Not all tests where validated.
I expect issues where urls are encoded and thus not replaced correctly.
2. I do not own any contents from wpt.live. All credit goes to all maintainers and contributors of [wpt](https://github.com/web-platform-tests/wpt)
