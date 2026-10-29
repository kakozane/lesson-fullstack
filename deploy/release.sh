#!/usr/bin/env bash
set -euo pipefail
# 只部署传入提交的镜像，数据库密码始终保存在服务器 .env 中。
sha="${1:?请传入提交 SHA}"
[[ "$sha" =~ ^[0-9a-f]{40}$ ]] || { echo '无效的提交 SHA' >&2; exit 1; }
cd /opt/lesson-fullstack
exec 9>.deploy.lock
flock -n 9 || { echo '已有部署正在执行' >&2; exit 1; }
test -f .env
export IMAGE_TAG="$sha"
compose=(docker compose --env-file .env -f "releases/$sha/compose.yaml")
"${compose[@]}" config --quiet
"${compose[@]}" pull
"${compose[@]}" up -d --wait --wait-timeout 120
curl --fail --silent --show-error http://127.0.0.1:6330/api/health
# 只在健康检查通过后记录成功版本；保留配置便于手动回滚。
cp "releases/$sha/compose.yaml" compose.yaml
sed -i "s/^IMAGE_TAG=.*/IMAGE_TAG=$sha/" .env
printf '%s\n' "$sha" > current-release
printf '\n部署完成：%s\n' "$sha"
