# Sub2API 服务器维护手册

最后更新：2026-07-28

本文档用于维护部署在 `192.220.25.216` 上的 Sub2API，以及本项目自定义的“上游余额”功能。文档不记录服务器密码、API Key、数据库密码或其他密钥。

## 1. 当前环境速查

| 项目 | 当前值 |
| --- | --- |
| 服务器 | `192.220.25.216` |
| 对外地址 | `http://192.220.25.216:8080` |
| 系统架构 | `x86_64` |
| Docker Compose | `v5.2.0` |
| 服务器源码目录 | `/opt/sub2api` |
| 部署目录 | `/opt/sub2api/deploy` |
| 基础 Compose | `/opt/sub2api/deploy/docker-compose.local.yml` |
| 自定义叠加 Compose | `/opt/sub2api/deploy/docker-compose.custom.yml` |
| 数据目录 | `/opt/sub2api/deploy/data` |
| PostgreSQL 数据目录 | `/opt/sub2api/deploy/postgres_data` |
| Redis 数据目录 | `/opt/sub2api/deploy/redis_data` |
| 自定义仓库 | `https://github.com/KKKK1723/sub2api` |
| 官方仓库 | `https://github.com/Wei-Shaw/sub2api` |
| 生产自定义分支 | `custom/upstream-balance` |
| 官方镜像同步分支 | `main` |

当前容器：

| 容器 | 用途 | 重启策略 |
| --- | --- | --- |
| `sub2api` | 应用服务，监听 `8080` | `unless-stopped` |
| `sub2api-postgres` | PostgreSQL | 由 Compose 管理 |
| `sub2api-redis` | Redis | 由 Compose 管理 |

服务器上的 `cli-proxy-api` 是另一个独立项目，维护 Sub2API 时不要停止、删除或重建它。

## 2. 当前生产版本

当前生产部署使用固定提交镜像：

```text
ghcr.io/kkkk1723/sub2api:custom-63f7fed8c1800065d5920ec068b79a0e38410ca0
```

镜像信息：

```text
版本：0.1.166-custom.63f7fed
提交：63f7fed8c1800065d5920ec068b79a0e38410ca0
镜像 ID：sha256:7f573e1ad869940eab0b427380b57091933375b71282fc7dafc1884834d063ac
```

首次上线前的完整备份：

```text
/opt/sub2api/deploy/backups/pre-custom-upstream-balance-20260728_102222
```

首次上线前的本地回滚镜像：

```text
local/sub2api-rollback:pre-custom-20260728_102222
```

上述信息是 2026-07-28 的快照。每次更新后，应同步修改本节。

## 3. 自定义功能说明

自定义功能为 API Key 类型账号增加上游余额查询能力：

- 添加或编辑 API Key 账号时，可以配置余额查询。
- 账号表格增加可选的“上游余额”列，默认隐藏。
- 支持单账号手动刷新、批量查询和后台定时刷新。
- 余额配置和快照保存在账号的 `extra` JSON 中，不需要数据库迁移。
- 查询复用账号本身的代理、TLS 指纹和并发控制设置。
- API Key 只在后端使用，不会返回给前端。
- 请求限制为同源地址，禁止重定向，超时 10 秒，响应最大 64 KB。

管理端接口：

```text
POST /api/v1/admin/accounts/:id/upstream-balance-probe
POST /api/v1/admin/accounts/upstream-balance-probe/batch
```

CC Switch 预设行为：

```http
GET {baseUrl}/v1/usage
Authorization: Bearer {apiKey}
```

兼容的常见返回字段：

```text
remaining
quota.remaining
balance
unit
quota.unit
is_active
isValid
```

使用时，在账号管理页面的列设置中启用“上游余额”，然后在对应 API Key 账号的编辑页面启用余额查询。已有账号不会被自动开启该功能。

## 4. 必须遵守的部署规则

### 4.1 永远同时使用两个 Compose 文件

所有会创建或重建 `sub2api` 容器的 Compose 命令，都必须同时包含：

```bash
-f docker-compose.local.yml -f docker-compose.custom.yml
```

正确示例：

```bash
cd /opt/sub2api/deploy
docker compose \
  -f docker-compose.local.yml \
  -f docker-compose.custom.yml \
  up -d --no-deps sub2api
```

禁止只运行下面的官方命令：

```bash
docker compose -f docker-compose.local.yml pull sub2api
docker compose -f docker-compose.local.yml up -d sub2api
```

只使用基础 Compose 文件会把应用重新切换到 `weishaw/sub2api:latest`，从而丢失自定义功能。

直接执行 `docker restart sub2api` 不会改变镜像，但它不能用于部署新版本。

### 4.2 生产环境使用不可变标签

生产服务器应使用：

```text
custom-<完整的 40 位 Git 提交 SHA>
```

不要在生产配置中长期使用移动标签 `custom-upstream-balance`。移动标签适合发现最新构建，不适合精确回滚。

### 4.3 不要清理未跟踪文件

服务器源码当前仍跟踪官方仓库的 `main` 分支，自定义镜像通过 Compose 叠加文件部署。以下内容是服务器上的未跟踪文件：

```text
deploy/docker-compose.custom.yml
deploy/backups/
```

禁止在 `/opt/sub2api` 中运行：

```bash
git clean -fd
git clean -fdx
```

这些命令可能删除叠加配置、备份，甚至被 Git 忽略的数据目录。普通 `git pull` 不会删除未跟踪的叠加文件，但更新后仍要确认文件存在。

## 5. 日常健康检查

```bash
cd /opt/sub2api/deploy

docker compose \
  -f docker-compose.local.yml \
  -f docker-compose.custom.yml \
  ps

curl -fsS http://127.0.0.1:8080/health

docker inspect sub2api --format \
  'status={{.State.Status}} health={{.State.Health.Status}} restarts={{.RestartCount}} image={{.Config.Image}}'

docker logs --since 10m --tail 200 sub2api

df -h /opt/sub2api/deploy
du -sh data postgres_data redis_data backups 2>/dev/null
```

正常状态应满足：

- `sub2api`、PostgreSQL、Redis 均为 `healthy`。
- `/health` 返回 `{"status":"ok"}`。
- `sub2api` 的镜像以 `ghcr.io/kkkk1723/sub2api:custom-` 开头。
- `RestartCount` 没有持续增长。

## 6. 合并官方更新

官方更新不能直接替换生产镜像。必须先把官方代码合并进自定义分支，保留余额功能并重新测试。

在维护机器的自定义仓库中执行：

```bash
git fetch origin
git fetch upstream

git switch main
git merge --ff-only upstream/main
git push origin main

git switch custom/upstream-balance
git merge main
```

如果发生冲突，重点检查账号实体、账号管理接口、路由、代理服务、账号编辑表单、账号表格列和中英文翻译。不能为了快速解决冲突而删除 `UpstreamBalance` 相关代码。

后端验证：

```bash
cd backend
go test ./internal/repository -count=1
go test ./internal/service -run UpstreamBalance -count=1
go test ./internal/handler/admin ./internal/server/routes -count=1
```

前端验证：

```bash
cd ../frontend
pnpm install --frozen-lockfile
pnpm run test:run
pnpm run lint:check
pnpm run typecheck
pnpm run build
```

提交前检查：

```bash
cd ..
git diff --check
git status
```

确认测试通过后推送：

```bash
git push origin custom/upstream-balance
```

GitHub Actions 会运行 `.github/workflows/custom-image.yml`，发布两个标签：

```text
ghcr.io/kkkk1723/sub2api:custom-upstream-balance
ghcr.io/kkkk1723/sub2api:custom-<完整提交 SHA>
```

查看构建结果：

```bash
gh run list \
  --repo KKKK1723/sub2api \
  --workflow custom-image.yml \
  --limit 10
```

推送只会构建镜像，不会自动更新服务器。只有构建成功后才能执行生产部署。

## 7. 每次部署前备份

下面的备份包含 `.env` 和数据库内容，目录权限必须保持为仅 root 可访问。

```bash
set -eu
cd /opt/sub2api/deploy
umask 077

ts=$(date +%Y%m%d_%H%M%S)
backup_dir="/opt/sub2api/deploy/backups/pre-update-$ts"
mkdir -p "$backup_dir"

cp -a \
  .env \
  docker-compose.local.yml \
  docker-compose.custom.yml \
  "$backup_dir/"

tar --exclude='data/logs' \
  -czf "$backup_dir/app-data.tar.gz" \
  data

docker inspect sub2api > "$backup_dir/sub2api-container-inspect.json"
docker image inspect "$(docker inspect sub2api --format '{{.Image}}')" \
  > "$backup_dir/sub2api-image-inspect.json"

db_user=$(docker inspect sub2api-postgres \
  --format '{{range .Config.Env}}{{println .}}{{end}}' \
  | sed -n 's/^POSTGRES_USER=//p' \
  | head -n 1)
[ -n "$db_user" ] || db_user=postgres

docker exec sub2api-postgres \
  pg_dumpall -U "$db_user" \
  | gzip -9 > "$backup_dir/postgres-all.sql.gz"

test -s "$backup_dir/postgres-all.sql.gz"

old_image=$(docker inspect sub2api --format '{{.Image}}')
docker image tag "$old_image" "local/sub2api-rollback:pre-update-$ts"

(
  cd "$backup_dir"
  sha256sum \
    .env \
    docker-compose.local.yml \
    docker-compose.custom.yml \
    app-data.tar.gz \
    sub2api-container-inspect.json \
    sub2api-image-inspect.json \
    postgres-all.sql.gz \
    > SHA256SUMS
  sha256sum -c SHA256SUMS
)

printf 'backup=%s\nrollback_image=%s\n' \
  "$backup_dir" \
  "local/sub2api-rollback:pre-update-$ts"
```

只有所有校验均显示 `OK`，才能继续部署。

## 8. 部署新镜像

先取得 GitHub Actions 成功构建对应的完整提交 SHA，然后将 `/opt/sub2api/deploy/docker-compose.custom.yml` 修改为：

```yaml
services:
  sub2api:
    image: ghcr.io/kkkk1723/sub2api:custom-<完整提交 SHA>
```

验证并部署：

```bash
set -eu
cd /opt/sub2api/deploy

docker pull ghcr.io/kkkk1723/sub2api:custom-<完整提交 SHA>

docker compose \
  -f docker-compose.local.yml \
  -f docker-compose.custom.yml \
  config -q

docker compose \
  -f docker-compose.local.yml \
  -f docker-compose.custom.yml \
  config --images

docker compose \
  -f docker-compose.local.yml \
  -f docker-compose.custom.yml \
  up -d --no-deps sub2api
```

`config --images` 的 `sub2api` 镜像必须是预期的 `ghcr.io/kkkk1723/sub2api:custom-<SHA>`。`--no-deps` 用于避免无必要地重启 PostgreSQL 和 Redis。

部署后验证：

```bash
docker inspect sub2api --format \
  'status={{.State.Status}} health={{.State.Health.Status}} restarts={{.RestartCount}} image={{.Config.Image}}'

curl -fsS http://127.0.0.1:8080/health
curl -fsS http://192.220.25.216:8080/health

docker logs --since 5m --tail 200 sub2api
```

还应登录管理后台进行以下人工检查：

1. 账号管理页面可以正常打开。
2. 列设置中存在“上游余额”。
3. API Key 账号编辑页面存在余额查询配置。
4. 使用测试账号执行一次余额刷新，结果和上游 `/v1/usage` 一致。
5. 现有网关请求仍能正常返回。

## 9. 回滚

应用异常时，优先只回滚应用镜像，不要立即恢复数据库。

将 `docker-compose.custom.yml` 中的镜像改为部署前生成的本地回滚标签，例如：

```yaml
services:
  sub2api:
    image: local/sub2api-rollback:pre-custom-20260728_102222
```

然后执行：

```bash
cd /opt/sub2api/deploy

docker compose \
  -f docker-compose.local.yml \
  -f docker-compose.custom.yml \
  config -q

docker compose \
  -f docker-compose.local.yml \
  -f docker-compose.custom.yml \
  up -d --no-deps sub2api

curl -fsS http://127.0.0.1:8080/health
docker logs --since 5m --tail 200 sub2api
```

当前余额功能没有数据库迁移，回滚当前版本通常不需要恢复数据库。未来如果官方更新引入数据库迁移，必须先确认新旧版本的数据库兼容性。数据库恢复会覆盖或冲突现有数据，只能在维护窗口中、再次备份当前状态后执行。

## 10. 常见问题

### 页面没有“上游余额”列

该列默认隐藏。在账号管理页面打开列设置并启用“上游余额”。如果仍不存在，检查浏览器是否缓存旧资源，并确认容器镜像是自定义 SHA 标签。

### 显示“未配置”或没有余额

确认以下项目：

- 账号类型是 API Key。
- 账号编辑页面已启用余额查询。
- CC Switch 站点实际支持 `GET /v1/usage`。
- `baseUrl` 正确，API Key 有权读取余额。
- 账号代理能够访问上游站点。

### 余额刷新失败

先查看应用日志，但不要把 API Key 输出到终端或工单：

```bash
docker logs --since 10m --tail 300 sub2api
```

常见原因包括上游接口不兼容、代理不可用、证书错误、超时、返回 JSON 字段不匹配或上游拒绝 API Key。

### 更新后功能消失

执行：

```bash
docker inspect sub2api --format '{{.Config.Image}}'
docker inspect sub2api --format \
  '{{index .Config.Labels "com.docker.compose.project.config_files"}}'
```

正确结果应为自定义镜像，并且配置文件列表同时包含：

```text
docker-compose.local.yml
docker-compose.custom.yml
```

如果镜像变成 `weishaw/sub2api:latest`，说明曾使用基础 Compose 文件单独重建。按“部署新镜像”章节重新部署即可。

### GHCR 镜像无法拉取

确认 GitHub Actions 构建成功、标签使用完整 SHA，并检查 GHCR package 是否保持公开。不要把个人 GitHub Token 写入 Compose 文件或提交到仓库。

## 11. 安全与维护纪律

- 不在仓库或本文档记录 root 密码、`.env` 内容、API Key、数据库密码。
- 优先使用 SSH 密钥登录，并定期轮换曾通过聊天或其他渠道传递的密码。
- `.env` 和备份目录只允许 root 读取。
- 部署前必做数据库备份，部署后必做健康检查。
- 不使用 `latest` 作为生产版本的唯一回滚依据。
- 不盲目运行官方安装脚本或 `docker-deploy.sh`，它们可能覆盖 Compose 配置。
- 不删除仍用于回滚的本地镜像和备份。
- 清理旧备份前，先确认至少保留一个已验证数据库备份和一个可启动的旧镜像。
- 修改防火墙、反向代理、PostgreSQL 或 Redis 前，应单独制定变更和回滚方案。

## 12. 一次标准更新的最短检查清单

1. 将官方 `upstream/main` 合并到 `custom/upstream-balance`。
2. 完成后端测试、前端测试、lint、类型检查和构建。
3. 推送自定义分支并等待 GitHub Actions 成功。
4. 记录准备部署的完整提交 SHA。
5. 在服务器生成数据库、配置和当前镜像备份，并验证校验和。
6. 将自定义 Compose 文件改为新的不可变 SHA 镜像。
7. 使用两个 Compose 文件执行 `config -q`。
8. 只重建 `sub2api`，不重启 PostgreSQL 和 Redis。
9. 检查容器健康、外部 `/health`、启动日志和现有 API 流量。
10. 登录后台人工验证上游余额功能。
11. 更新本文档中的当前版本、备份目录和回滚镜像。
