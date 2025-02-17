# traefik-provider-frp

> Provides frps proxy information to Traefik, enabling automatic reverse proxying

frp is the most mainstream solution for intranet penetration. In common use cases, frps is deployed on a server with a public IP, and frpc is deployed on an intranet server. After manual configuration, the HTTP service of the intranet server can be directly accessed on the specified port of the public server.

When HTTPS access to the service is required, frp becomes inadequate. Although this requirement can be achieved through <https://gofrp.org/zh-cn/docs/examples/https2http/>, it requires placing certificate files on the client side, posing a risk of certificate leakage and making automatic rotation difficult. A better implementation is to deploy a reverse proxy service like Traefik on the public server, achieving secure service access through virtual hosts and wildcard certificates at the gateway, and solving the certificate rotation problem through built-in ACME. However, when the number of services to be proxied is large, manually modifying Traefik configurations becomes cumbersome.

traefik-provider-frp can effectively solve this problem. traefik-provider-frp retrieves proxy information through the frps admin API and can be configured as an HTTP Provider for Traefik, automatically providing proxy information to Traefik.

You can leverage <https://github.com/117503445/frpc-controller> to automatically generate frpc configuration files based on Docker Labels, further enhancing automation.

## Quick Start

Prepare configuration files and Docker Compose declaration files

```sh
git clone https://github.com/117503445/traefik-provider-frp.git
cd traefik-provider-frp/docs/example
```

Start the services

```sh
docker compose up -d
```

Verify

```sh
curl -H "Host: app1.example.com" 127.0.0.1:80
# show output of app1

curl -H "Host: app2.example.com" http://127.0.0.1:80
# show output of app2
```

The following is a detailed explanation of this example

`app1` and `app2` are two simple HTTP services, assuming they are intranet services that need to be proxied

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

`frpc` is also deployed on the intranet server

```yaml
services:
  frpc:
    image: snowdreamtech/frpc
    restart: unless-stopped
    volumes:
      - ./config/frpc/frpc.toml:/etc/frp/frpc.toml
```

The following is the configuration for frpc. In actual configurations, you can use Traefik to wrap a layer of tls around the bindPort of frps, and then have frpc connect to frps using the wss protocol. This is because frp uses the TCP protocol by default, and if security is to be ensured, you need to distribute certificates yourself, which is very troublesome to maintain; using websocket (wss) can reuse the reverse proxy infrastructure, using trusted certificates instead of self-signed certificates.

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

`Traefik` is deployed on the public server

```yaml
services:
  traefik:
    image: traefik
    volumes:
      - ./config/traefik/traefik.yml:/etc/traefik/traefik.yaml
    ports:
      - "80:80"
```

The following is the configuration for Traefik. Use traefik-provider-frp as an HTTP Provider.

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

`frps` is also deployed on the public server

```yaml
services:
  frps:
    image: snowdreamtech/frps
    restart: unless-stopped
    volumes:
      - ./config/frps/frps.toml:/etc/frp/frps.toml
```

The following is the configuration for frps. Configure traefik-provider-frp as a plugin, so that traefik-provider-frp is notified every time frps adds or deletes a proxy, avoiding the need for traefik-provider-frp to poll frps.

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

`traefik-provider-frp` is deployed on the public server

```yaml
services:
  traefik-provider-frp:
    image: 117503445/traefik-provider-frp
    volumes:
      - ./config/traefik-provider-frp/config.toml:/workspace/config.toml
      - ./config/traefik-provider-frp/traefik.tmpl:/workspace/traefik.tmpl
```

The following is the configuration for traefik-provider-frp. `config.toml` defines the address of frps-admin.

```toml
# ./config/traefik-provider-frp/config.toml

"frps-admin.base-url" = "http://frps:7500"
"frps-admin.username" = "admin"
"frps-admin.password" = "12345678"
# "traefik-writer.output-path" = "/shared-dynamic-cfg/frp_dynamic.yml"
```

`traefik.tmpl` is the template to be provided to Traefik. You can refer to <https://masterminds.github.io/sprig/>. Where `$SERVICES$` will be replaced with frps proxy information, the example is `[{"service":"app1","url":"http://frps:2131"},{"service":"app2","url":"http://frps:6432"}]`

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

## Configuration Reference

The file `/workspace/config.toml` inside the container is the configuration file

| Configuration Item | Description | Default Value |
| --- | --- | --- |
| `frps-admin.base-url` | The address of frps-admin | `http://frps:7500` |
| `frps-admin.username` | The username of frps-admin | "" |
| `frps-admin.password` | The password of frps-admin | "" |

The file `/workspace/traefik.tmpl` inside the container is the template file, you can refer to <https://doc.traefik.io/traefik/providers/file/#go-templating> and <https://masterminds.github.io/sprig/>. Since Traefik's Template only supports File Provider, and traefik-provider-frp is an HTTP Provider, to support Template, template execution parsing can only be done within traefik-provider-frp. When writing templates, be cautious about handling file-related issues.

## Implementation

When frps adds or deletes a proxy, it notifies traefik-provider-frp. Upon receiving the notification, traefik-provider-frp retrieves the full proxy information through the frps-admin API, combines it with `/workspace/traefik.tmpl`, generates dynamic configurations, and provides them to Traefik as an HTTP Provider.
