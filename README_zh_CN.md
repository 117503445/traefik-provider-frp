# traefik-provider-frp

> 将 frps 代理信息提供给 Traefik，从而实现自动反向代理

frp 是最主流的内网穿透方案。在常见的使用案例中，将 frps 部署在具有公网 IP 的服务器上，并将 frpc 部署在内网服务器上，进行手动配置后可以在公网服务器的指定端口直接访问到内网服务器的 HTTP 服务。

当需要使用 HTTPS 访问服务时，frp 就难以为继了。尽管可以通过 <https://gofrp.org/zh-cn/docs/examples/https2http/> 实现此需求，但是需要在客户端放置证书文件，存在证书泄漏风险，且难以实现自动轮换。更好的实现方式是在公网服务器上部署 Traefik 等反向代理服务，在网关中通过虚拟主机、泛域名证书实现安全的服务访问，并通过内置的 ACME 解决证书轮换问题。然而，当需要代理的服务数量较多时，手动修改 Traefik 配置就会十分麻烦。

traefik-provider-frp 可以很好的解决这个问题。traefik-provider-frp 通过 frps admin API 获取代理信息，并可以配置为 Traefik 的 HTTP Provider，从而自动为 Traefik 提供代理信息。

可以借助 <https://github.com/117503445/frpc-controller> 根据 Docker Label 自动生成 frpc 配置文件，进一步提升自动化。

## 快速开始

准备配置文件和 Docker Compose 声明文件

```sh
git clone https://github.com/117503445/traefik-provider-frp.git
cd traefik-provider-frp/docs/example
```

启动服务

```sh
docker compose up -d
```

验证

```sh
curl -H "Host: app1.example.com" 127.0.0.1:80
# show output of app1

curl -H "Host: app2.example.com" http://127.0.0.1:80
# show output of app2
```

以下为此示例的详细说明

`app1` 和 `app2` 是两个简单的 HTTP 服务，假设它们是需要被代理的内网服务

```yaml
services:
  app1:
    image: traefik/whoami
    hostname: app1
    restart: unless-stopped
  app2:
    image: traefik/whoami
    hostname: app2
    restart: unless-stopped
```

`frpc` 也部署在内网服务器上

```yaml
services:
  frpc:
    image: snowdreamtech/frpc
    restart: unless-stopped
    volumes:
      - ./config/frpc/frpc.toml:/etc/frp/frpc.toml
```

以下为 frpc 的配置。在实际的配置中，可以使用 Traefik 为 frps 的 bindPort 套一层 tls，然后 frpc 使用 wss 协议连接 frps。这是因为 frp 默认使用 TCP 协议，如果要确保安全，就需要自己分发证书，运维非常麻烦；而使用 websocket(wss) 可以复用反向代理基础设施，使用可信任的证书而不是自签名证书。

```toml
# ./config/frpc/frpc.toml
loginFailExit = false

serverAddr = "frps"
transport.protocol = "websocket"

auth.token = "123456"

[[proxies]]
name = "app1"
type = "tcp"
localIP = "app1"
localPort = 80
metadatas.domain = ""

[[proxies]]
name = "app2"
type = "tcp"
localIP = "app2"
localPort = 80
metadatas.domain = ""
```

`Traefik` 部署在公网服务器上

```yaml
services:
  traefik:
    image: traefik
    volumes:
      - ./config/traefik/traefik.yml:/etc/traefik/traefik.yaml
    ports:
      - "80:80"
```

以下为 Traefik 的配置。使用 traefik-provider-frp 作为 HTTP Provider。

```yaml
# ./config/traefik/traefik.yml
log:
  level: DEBUG

providers:
  http:
    endpoint: "http://traefik-provider-frp:8081/traefik"

api:
  dashboard: true
  # insecure: true

entryPoints:
  external:
    address: ":80"
    http3: {}
```

`frps` 也部署在公网服务器上

```yaml
services:
  frps:
    image: snowdreamtech/frps
    restart: unless-stopped
    volumes:
      - ./config/frps/frps.toml:/etc/frp/frps.toml
```

以下为 frps 的配置。将 traefik-provider-frp 配置为插件，每次 frps 增删代理的时候都会通知 traefik-provider-frp，避免 traefik-provider-frp 轮询 frps。

```toml
# ./config/frps/frps.toml
bindPort = 80
auth.token = "123456"

[webServer]
addr = "0.0.0.0"
port = 7500
user = "admin"
password = "12345678"

[[httpPlugins]]
name = "manager"
addr = "http://traefik-provider-frp:8021"
path = "/frp"
ops = ["NewProxy", "CloseProxy"]
```

`traefik-provider-frp` 部署在公网服务器上

```yaml
services:
  traefik-provider-frp:
    image: 117503445/traefik-provider-frp
    volumes:
      - ./config/traefik-provider-frp/config.toml:/workspace/config.toml
      - ./config/traefik-provider-frp/traefik.tmpl:/workspace/traefik.tmpl
```

以下为 traefik-provider-frp 的配置。`config.toml` 定义了 frps-admin 的地址。

```toml
# ./config/traefik-provider-frp/config.toml

"frps-admin.base-url" = "http://frps:7500"
"frps-admin.username" = "admin"
"frps-admin.password" = "12345678"
# "traefik-writer.output-path" = "/shared-dynamic-cfg/frp_dynamic.yml"
```

`traefik.tmpl` 是要提供给 Traefik 的模板。可以参考 <https://masterminds.github.io/sprig/>。其中 `$SERVICES$` 会被替换为 frps 代理信息，例子为 `[{"service":"app1","url":"http://frps:2131"},{"service":"app2","url":"http://frps:6432"}]`

```yaml
# ./config/traefik-provider-frp/traefik.tmpl

{{- $data := `
   $SERVICES$
` | fromJson -}}

http:
  routers:
    {{- range $index, $element := $data }}
    {{ $element.service }}:
      rule: Host(`{{ $element.service }}.example.com`)
      entryPoints:
        - external
      service: {{ $element.service }}
    {{- end }}
  services:
    {{- range $index, $element := $data }}
    {{ $element.service }}:
      loadBalancer:
        servers:
        - url: {{ $element.url }}
    {{- end }}
```

## 配置参考

容器内 `/workspace/config.toml` 是配置文件

| 配置项 | 描述 | 默认值 |
| --- | --- | --- |
| `frps-admin.base-url` | frps-admin 的地址 | `http://frps:7500` |
| `frps-admin.username` | frps-admin 的用户名 | "" |
| `frps-admin.password` | frps-admin 的密码 | "" |

容器内 `/workspace/traefik.tmpl` 是模板文件，可以参考 <https://doc.traefik.io/traefik/providers/file/#go-templating> 和 <https://masterminds.github.io/sprig/>。因为 Traefik 的 Template 只支持 File Provider，而 traefik-provider-frp 是一个 HTTP Provider，所以为了支持 Template，只能在 traefik-provider-frp 中进行模板执行解析。编写模板时，应谨慎处理和文件相关的问题。

## 实现

frps 增删代理的时候，会通知 traefik-provider-frp。traefik-provider-frp 收到通知后，通过 frps-admin API 获取全量代理信息，并与 `/workspace/traefik.tmpl` 结合，生成动态配置，并以 HTTP Provider 的方式提供给 Traefik。
