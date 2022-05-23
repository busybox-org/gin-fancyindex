# GIN-FancyIndex

Develop a compact and lightweight file directory index server based on gin
+ Inspired by [caddy](https://caddyserver.com/) [file_server](https://caddyserver.com/docs/caddyfile/directives/file_server)

## Parameters
+ `PORT`: the port to listen on. default: 8080
+ `RELATIVE_PATH`: the path to serve. default: `/`
+ `ROOT`: the root directory to serve. default: `/share`
+ `AUTH`: enable authentication. default: false
+ `AUTH_USER`: the username to use for authentication. default: `admin`
+ `AUTH_PASS`: the password to use for authentication. default: `admin`


## RUN

```bash
docker pull xmapst/gin-fancyindex
docker run -it -p 8080:8080 -v /share:/share xmapst/gin-fancyindex:latest
```

browser open [http://localhost:8080](http://localhost:8080)

![gin-fancyindex](https://raw.githubusercontent.com/xmapst/gin-fancyindex/main/gin-fancyindex.jpg)

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fxmapst%2Fgin-fancyindex.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fxmapst%2Fgin-fancyindex?ref=badge_shield)


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fxmapst%2Fgin-fancyindex.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fxmapst%2Fgin-fancyindex?ref=badge_large)