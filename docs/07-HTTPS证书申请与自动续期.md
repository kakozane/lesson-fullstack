# Let’s Encrypt 与 acme.sh HTTPS 证书申请和自动续期

> 文中的 `203.0.113.10` 仅为文档示例 IP，并非课程服务器地址；实际部署时请替换为自己的服务器公网 IP。

本项目使用 Let’s Encrypt 签发免费 HTTPS 证书，通过 acme.sh 自动完成申请、续签和安装，由宿主机 Nginx 提供 HTTPS。以 `test.retniw.cc` 为例，了解从 DNS 解析到浏览器安全访问的完整过程，以及日常需要保持哪些配置。

首次配置完成后，通常不需要定期手动下载、上传证书。域名有效、解析正确、服务器和续期任务正常，是这套自动化持续运行的前提。新增另一个域名时，还需要为它配置站点和首次申请证书，仅增加 DNS 记录不会自动完成这些工作。

本文中的服务器路径对应 2026 年 10 月 29 日部署的课程环境。应用发布过程见 [线上部署与自动发布](06-线上部署与自动发布.md)。

## 1. 各部分分别负责什么

| 部分 | 职责 | 当前项目 |
|---|---|---|
| DNS 服务 | 将域名解析到服务器地址 | 阿里云 DNS，`test.retniw.cc` 指向 `203.0.113.10` |
| Let’s Encrypt | 作为证书颁发机构，验证域名控制权并签发证书 | 签发 `test.retniw.cc` 的证书 |
| acme.sh | 通过 ACME 协议与签发机构交互，申请、续签和安装证书 | 运行在服务器宿主机上 |
| Nginx | 使用证书提供 HTTPS，并代理业务请求 | 宿主机监听 80、443 端口 |
| cron | 定期执行续期检查 | 每天 22:00，按服务器 Asia/Shanghai 时区执行 |

ACME 是客户端与证书颁发机构进行自动化交互的协议。Let’s Encrypt 是签发机构，acme.sh 是客户端工具；acme.sh 也支持其他签发机构，因此本项目申请时显式指定 `--server letsencrypt`。[acme.sh 官方说明](https://github.com/acmesh-official/acme.sh)

域名注册商、DNS 服务商和证书签发机构可以分别选择。本方案只使用阿里云的域名解析，不需要从阿里云控制台领取证书，也没有配置阿里云 DNS AccessKey。

## 2. DNS 与域名验证的关系

在 `retniw.cc` 的 DNS 控制台中配置：

| 记录类型 | 主机记录 | 记录值 |
|---|---|---|
| A | `test` | `203.0.113.10` |

DNS 解决的是“这个域名去哪里访问”。申请证书还需要证明申请者能控制该域名对应的服务。

本项目采用 **HTTP-01 验证**：

1. acme.sh 向 Let’s Encrypt 提交域名。
2. 签发机构给出验证要求，acme.sh 生成相应的临时验证文件。
3. Nginx 通过公开 URL 提供该文件。
4. Let’s Encrypt 解析域名，访问文件并核对内容。
5. 验证通过后签发证书，acme.sh 下载并安装。

验证地址形如：

```text
http://test.retniw.cc/.well-known/acme-challenge/<临时标识>
```

HTTP-01 验证要求公网可以访问 **80 端口**。本项目让这个路径直接返回文件，其他 HTTP 请求跳转 HTTPS。无需为每次续签修改 DNS，也无需停掉正在运行的 Nginx。[Let’s Encrypt 验证机制](https://letsencrypt.org/docs/challenge-types/)

## 3. HTTPS 请求与验证请求走不同路径

正常业务访问：

```text
浏览器
  → DNS 找到 203.0.113.10
  → 宿主机 Nginx 的 443 端口，建立 HTTPS 连接
  → 127.0.0.1:6330
  → Web 容器：返回完整版前端，或代理 /api 请求
  → Gin API 容器
```

证书验证：

```text
Let’s Encrypt 验证服务器
  → DNS 找到 203.0.113.10
  → 宿主机 Nginx 的 80 端口
  → /.well-known/acme-challenge/ 下的临时文件
```

证书验证不经过前端或 Go 容器。因此，通过 GitHub Actions 重新发布应用、重建业务容器，不会删除证书配置。即使 Go 服务暂时停止，只要宿主机 Nginx 和验证目录正常，HTTP 域名验证仍然可以完成。

## 4. 当前服务器上的文件

| 路径 | 用途 |
|---|---|
| `/opt/lesson-fullstack/acme/acme.sh` | ACME 客户端脚本 |
| `/opt/lesson-fullstack/acme/` | 客户端、账户信息、域名配置和内部证书文件 |
| `/www/wwwroot/test.retniw.cc/.well-known/acme-challenge/` | 临时验证文件目录 |
| `/www/server/panel/vhost/nginx/test.retniw.cc.conf` | 宿主机域名配置 |
| `/www/server/panel/vhost/cert/test.retniw.cc/fullchain.pem` | 安装给 Nginx 使用的完整证书链 |
| `/www/server/panel/vhost/cert/test.retniw.cc/privkey.pem` | 证书私钥 |
| `/etc/cron.d/lesson-fullstack-cert` | 每日续期检查任务 |
| `/opt/lesson-fullstack/acme/renew.log` | 定时续期日志 |

这些是服务器上的路径。仓库中的 `deploy/site.conf` 是宿主机 Nginx 配置模板；`deploy/nginx.conf` 是 Web 容器内部的配置，二者职责不同。

Nginx 引用的是安装后的证书文件，不直接引用 acme.sh 内部工作目录中的证书。安装命令会记录目标路径与重载命令，后续续签沿用这些设置。备份时应保护账户与私钥文件，不能把它们提交到 GitHub。

## 5. 首次配置的顺序

以下命令用于理解和复现首次部署，均在服务器上以具备相应权限的账号执行。当前课程服务器已经完成配置，无需重新执行。换服务器或域名时，需要同步修改路径、域名和 Nginx 上游地址。

### 5.1 准备解析与 HTTP 验证入口

服务器需具备 Nginx、curl、OpenSSL 和 cron。确认 DNS 已指向服务器，安全组和防火墙允许公网访问 TCP 80、443。

```sh
install -d /www/wwwroot/test.retniw.cc/.well-known/acme-challenge
```

首次签发前，站点先使用 HTTP 配置；证书尚不存在时，不要直接启用引用证书文件的 HTTPS 配置：

```nginx
server {
    listen 80;
    server_name test.retniw.cc;

    location ^~ /.well-known/acme-challenge/ {
        root /www/wwwroot/test.retniw.cc;
        try_files $uri =404;
    }

    location / {
        proxy_pass http://127.0.0.1:6330;
    }
}
```

将配置保存到本机 Nginx 已包含的站点目录。当前服务器使用 `/www/server/panel/vhost/nginx/test.retniw.cc.conf`。然后执行：

```sh
nginx -t && nginx -s reload
```

可以用一个临时文件验证路径映射：

```sh
printf 'acme-webroot-ok\n' > /www/wwwroot/test.retniw.cc/.well-known/acme-challenge/lesson-check.txt
```

从服务器以外的电脑访问，确认能返回 `acme-webroot-ok`：

```sh
curl -f http://test.retniw.cc/.well-known/acme-challenge/lesson-check.txt
```

验证完成后，在服务器删除本次测试文件：

```sh
rm /www/wwwroot/test.retniw.cc/.well-known/acme-challenge/lesson-check.txt
```

### 5.2 准备 acme.sh 并申请证书

本项目将脚本单独保存到 `/opt/lesson-fullstack/acme`，没有使用默认的 `~/.acme.sh`。以下下载地址固定到首次部署使用的官方源码提交，便于复现：

```sh
install -d -m 700 /opt/lesson-fullstack/acme
curl -fL \
  https://raw.githubusercontent.com/acmesh-official/acme.sh/37d064f7b580055d3f2ddf9a9c960f754f3694d6/acme.sh \
  -o /opt/lesson-fullstack/acme/acme.sh
chmod 700 /opt/lesson-fullstack/acme/acme.sh
```

申请指定域名的证书：

```sh
sh /opt/lesson-fullstack/acme/acme.sh \
  --home /opt/lesson-fullstack/acme \
  --issue \
  --server letsencrypt \
  --webroot /www/wwwroot/test.retniw.cc \
  -d test.retniw.cc
```

`--webroot` 是网站根目录，acme.sh 会在它下面创建 `.well-known/acme-challenge/` 中的文件。不要把 `--webroot` 填成最终的 challenge 子目录。

这里采用保存脚本、显式指定工作目录、单独配置 cron 的方式。如果改用官方安装器，需要检查其是否已创建定时任务，避免重复配置。[官方安装说明](https://github.com/acmesh-official/acme.sh/wiki/How-to-install)

### 5.3 安装证书并启用 HTTPS

当前项目使用 ECC 证书，安装时带 `--ecc`：

```sh
install -d -m 700 /www/server/panel/vhost/cert/test.retniw.cc

sh /opt/lesson-fullstack/acme/acme.sh \
  --home /opt/lesson-fullstack/acme \
  --install-cert \
  -d test.retniw.cc \
  --ecc \
  --key-file /www/server/panel/vhost/cert/test.retniw.cc/privkey.pem \
  --fullchain-file /www/server/panel/vhost/cert/test.retniw.cc/fullchain.pem \
  --reloadcmd 'nginx -t && nginx -s reload'
```

证书安装成功后，使用仓库 `deploy/site.conf` 的完整配置替换首次申请时的 HTTP 配置。它保留 80 端口的验证路径，并增加 HTTPS 服务。部署模板到上述宿主机站点路径后，执行：

```sh
nginx -t && nginx -s reload
curl -I https://test.retniw.cc
```

`--reloadcmd` 很关键：续签得到新文件后，Nginx 还需要重载，才能让新连接使用新证书。

### 5.4 配置每日检查

当前 `/etc/cron.d/lesson-fullstack-cert` 内容如下：

```cron
SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
0 22 * * * root sh /opt/lesson-fullstack/acme/acme.sh --cron --home /opt/lesson-fullstack/acme >> /opt/lesson-fullstack/acme/renew.log 2>&1
```

该文件由 root 拥有、权限为 `0644`，最后一行保留换行。`/etc/cron.d` 的任务行包含执行用户 `root`；如果使用 `crontab -e` 编辑用户自己的任务，格式中不应再写这个用户字段。

当前服务器的 cron 服务名为 `crond`，状态应为 `active`：

```sh
systemctl is-active crond
```

其他发行版可能使用 `cron` 服务名，应以实际系统为准。

## 6. 自动续期时会发生什么

服务器凌晨会关机，因此检查安排在晚间；请确保每天 22:00 服务器处于运行状态，停机期间错过的普通 cron 任务不会自动补跑。

每天北京时间 22:00 的任务会检查证书是否需要续签：未到续期时间就跳过；需要续签时申请新证书，并在需要域名验证时通过保留的 HTTP 路径完成验证，再安装文件和重载 Nginx。每天检查不等于每天签发新证书。[acme.sh 自动化流程](https://github.com/acmesh-official/acme.sh)

当前签发机构的默认有效期为 90 天，但应以实际证书与客户端记录为准，不要把“每 90 天执行一次”写死到任务中。有效期政策未来可能调整，客户端也需要按官方版本维护。[Let’s Encrypt 有效期说明](https://letsencrypt.org/docs/cert-lifetimes/)

短暂网络故障不会立即使原有证书失效。修复后，后续定时任务仍有机会完成续签；如果故障一直持续到当前证书过期，浏览器才会因过期证书报错。

本项目目前保留续期日志，没有配置续期失败或临期通知。固定版本的 acme.sh 也没有开启自动升级；长期使用需要关注客户端更新和证书有效期。

## 7. 日常检查与常见问题

查看本地安装的证书有效期：

```sh
openssl x509 \
  -in /www/server/panel/vhost/cert/test.retniw.cc/fullchain.pem \
  -noout -issuer -dates
```

查看公网 Nginx 实际提供的证书，排除“文件已更新，但服务未重载”的情况：

```sh
openssl s_client -connect test.retniw.cc:443 -servername test.retniw.cc </dev/null 2>/dev/null \
  | openssl x509 -noout -issuer -dates
```

查看日志和任务：

```sh
tail -n 50 /opt/lesson-fullstack/acme/renew.log
cat /etc/cron.d/lesson-fullstack-cert
systemctl is-active crond
nginx -t
```

需要手动运行一次正常检查时：

```sh
sh /opt/lesson-fullstack/acme/acme.sh --cron --home /opt/lesson-fullstack/acme
```

该命令不是只读模拟；到达续期条件时，会实际续签、安装并重载。不要为日常检查反复加 `--force`，重复签发可能触发签发机构的频率限制。[Let’s Encrypt 签发限制](https://letsencrypt.org/docs/rate-limits/)

| 现象或变更 | 排查重点 |
|---|---|
| HTTP 验证返回 404 | 确认 DNS、Nginx 命中的站点和 webroot 路径一致；没有验证文件时，访问该目录下随意编造的文件名返回 404 是正常的 |
| HTTP 验证超时 | 检查公网 80 端口、安全组、防火墙和 Nginx 状态 |
| 网站能访问，但续签失败 | 检查验证路径是否被鉴权、CDN 或 WAF 拦截，服务器能否访问签发机构，以及续期日志 |
| 文件中的有效期已更新，浏览器仍显示旧证书 | 检查 Nginx 引用路径、重载结果；若增加了 CDN，还应检查 CDN 对外提供的证书 |
| 更换服务器 IP | 修改 DNS，并将站点、验证目录、ACME 配置与证书续期安排迁移到新服务器 |
| 增加 AAAA 记录或多个 IP | 确保所有对外地址都能正确提供验证文件，不要保留不可达或指向错误服务器的记录 |
| 域名仍有效，但证书过期 | 检查 cron 是否运行、磁盘与权限是否正常，以及最近一次成功续签的日志 |

## 8. 新增域名与通配符证书

新增 `demo.retniw.cc` 时，需要新增 DNS 记录、配置对应的 Nginx 站点、为新域名申请并安装证书。使用同一 acme.sh 工作目录管理时，已有的每日检查任务可以处理其中已配置的域名，不必每个域名再创建一条相同 cron。

当前证书只覆盖 `test.retniw.cc`，不会自动覆盖 `admin.retniw.cc` 或其他子域名。

如果希望一张证书覆盖 `*.retniw.cc`，需要使用 DNS-01 等支持通配符的验证方式。自动化的 DNS-01 通常通过 DNS 服务商 API 更新 `_acme-challenge` TXT 记录，这时才涉及 DNS API 凭据；它不是当前部署使用的方式。[域名验证方式比较](https://letsencrypt.org/docs/challenge-types/)
