# GIN-FancyIndex

Develop a compact and lightweight file directory index server based on gin
+ Inspired by [caddy](https://caddyserver.com/) [file_server](https://caddyserver.com/docs/caddyfile/directives/file_server)

## Parameters
+ `PORT`: the port to listen on. default: 8080
+ `ROOT`: the root directory to serve. default: `/share`
+ `AUTH`: enable authentication. default: false
+ `USER`: the username to use for authentication. default: `admin`
+ `PASS`: the password to use for authentication. default: `admin`


## RUN

```bash
docker pull xmapst/gin-fancyindex
docker run -it -p 8080:8080 -v /share:/share xmapst/gin-fancyindex:latest
```

browser open [http://localhost:8080](http://localhost:8080)

![gin-fancyindex](https://raw.githubusercontent.com/xmapst/gin-fancyindex/main/gin-fancyindex.jpg)
