# Custom Upstream Balance Branch

This fork keeps the upstream project and the local balance feature on separate
branches:

- `main`: mirror of `Wei-Shaw/sub2api:main`
- `custom/upstream-balance`: production branch containing the local feature

Do not deploy `weishaw/sub2api:latest`. The custom branch publishes:

- `ghcr.io/kkkk1723/sub2api:custom-upstream-balance`
- `ghcr.io/kkkk1723/sub2api:custom-<full-commit-sha>`

## Update From Upstream

1. Sync the fork's `main` branch with the official repository:

   ```bash
   gh repo sync KKKK1723/sub2api --branch main
   ```

2. Merge the updated mirror into the custom branch:

   ```bash
   git fetch origin upstream
   git switch custom/upstream-balance
   git merge origin/main
   ```

3. Resolve conflicts without dropping the balance commits, then run:

   ```bash
   cd backend
   go test ./internal/repository ./internal/handler/admin ./internal/server/routes
   go test ./internal/service -run UpstreamBalance
   cd ../frontend
   pnpm install --frozen-lockfile
   pnpm run lint:check
   pnpm run typecheck
   pnpm run build
   ```

4. Push the custom branch. GitHub Actions builds and publishes the image:

   ```bash
   git push origin custom/upstream-balance
   ```

## Deploy

Use the compose overlay so upstream compose updates cannot switch the service
back to the official image:

```bash
docker compose \
  -f deploy/docker-compose.yml \
  -f deploy/docker-compose.custom.yml \
  pull sub2api
docker compose \
  -f deploy/docker-compose.yml \
  -f deploy/docker-compose.custom.yml \
  up -d sub2api
```

Keep database and compose backups before every update.

## Roll Back

Change `deploy/docker-compose.custom.yml` from the moving tag to a previously
published immutable tag such as `custom-<full-commit-sha>`, then run the same
`docker compose up -d sub2api` command.
