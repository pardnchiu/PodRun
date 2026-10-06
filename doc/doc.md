# PodRun - Documentation

Last updated: 2026-10-06

> Back to [README](../README.md)

## Prerequisites

| Requirement | Where | Notes |
|---|---|---|
| Go 1.25.1+ | Build host | Taken from `go.mod` |
| C compiler (`CGO_ENABLED=1`) | Build host (API server) | `github.com/mattn/go-sqlite3` requires cgo |
| `sshpass`, `rsync`, `ssh`, `curl`, `unzip` | Local (CLI) | Installed automatically on startup when missing: `brew` on macOS; `apt` / `dnf` / `yum` / `pacman` in that order on Linux |
| macOS or Linux | Local (CLI) | Other operating systems abort when packages are missing |
| SSH password login | Remote host | The CLI connects through `sshpass` |
| Podman + `podman compose` | Remote host | Container runtime |
| Write access to `/home/podrun` | Remote host | Remote project folders are always created under this path |

## Installation

### From Source

```bash
git clone https://github.com/pardnchiu/PodRun.git
cd PodRun
go build -o podrun ./cmd/cli
CGO_ENABLED=1 go build -o podrun-api ./cmd/api
```

### Using go install (CLI only)

```bash
go install github.com/pardnchiu/PodRun/cmd/cli@latest
```

`go install` names the binary `cli` after the package directory; rename it to `podrun` if preferred.

## Configuration

### Environment Variables

Both the CLI and the API server load `.env` from the **current working directory** via `godotenv`; a missing file only logs a warning and system environment variables still apply.

| Variable | Used by | Required | Default | Description |
|---|---|---|---|---|
| `PODRUN_SERVER` | CLI | Yes | — | Remote hostname or IP |
| `PODRUN_USERNAME` | CLI | Yes | — | Remote SSH user |
| `PODRUN_PASSWORD` | CLI | Yes | — | Remote SSH password (passed to `sshpass`) |
| `DB_PATH` | API server | No | Inside a container (`/.dockerenv` exists): `/data/database.db`; otherwise `~/.podrun/database.db` | SQLite file path; `~/.podrun/` is created automatically for the default path |
| `ALLOW_EMAILS` | API server | No | — | Listed in `.env.expample`; currently disabled in code |

### Config File

```bash
cp .env.expample .env
```

```dotenv
PODRUN_SERVER=192.168.1.100
PODRUN_USERNAME=podrun
PODRUN_PASSWORD=yourpassword
```

## Usage

### Start the API Server

The API server reads `sql/create.sql` through a relative path to create tables, so run it from the repo root. The CLI always targets `localhost:8080`, so both must run on the same machine.

```bash
cd PodRun
./podrun-api
```

### Basic: Deploy the Current Project

Run inside a project folder that contains `docker-compose.yml` or `docker-compose.yaml`:

```bash
cd ~/projects/my-app
podrun up -d
```

`up` performs these steps:

1. Tests the SSH connection and aborts on failure
2. Creates `/home/podrun/<folder-name>_<first-8-hash-chars>/` on the remote host
3. When the remote folder is not empty, runs an rsync dry-run and asks for `y` before applying changes; then syncs with `rsync -avz --delete` (excluding `node_modules/`, `vendor/`, `.git/`, `.venv/`, `.next/`, `*.log`, and more)
4. Copies the compose file to `docker-compose.podrun.yml`, strips host port bindings (`8080:80` → `80`), and appends `:z` to volumes starting with `./`
5. Removes old containers with `down -v`, then runs `podman compose -f docker-compose.podrun.yml up -d`
6. Looks up the Pod ID and name, and prints container port mappings in `-d` mode
7. Sends the deployment info and an operation record to the API server

Without `-d`, `up` runs in the foreground; pressing `Ctrl+C` triggers `podman compose down` on the remote host.

### Advanced: Choose the Project Path and Compose File

```bash
# Set the local project folder
podrun up -d --folder=/path/to/project

# Pass the folder positionally (must start with ./ or / and exist)
podrun up -d ./services/api

# Set the compose file; without --folder, its directory becomes the project folder
podrun up -d -f ./services/api/docker-compose.yml

# Override the auto-generated deployment UID
podrun up -d -u my-app-prod
```

### Advanced: Operations

```bash
# Show container status
podrun ps

# Follow logs (-f after logs means follow, not file)
podrun logs -f

# Open a shell in a container
podrun exec web sh

# Stop and remove containers
podrun down

# Full cleanup: containers, volumes, images, and the remote project folder
podrun clear
```

`clear` deletes the remote folder through `podman run --privileged alpine`, which removes files created by rootless containers that the SSH user cannot delete directly.

## CLI Reference

### Commands

| Command | Syntax | Description |
|---|---|---|
| `up` | `podrun up [-d] [flags]` | Sync, rewrite compose, recreate and start containers, record the deployment |
| `clear` | `podrun clear` | `down -v`, `down --rmi all`, then delete the remote project folder |
| `down` | `podrun down [args]` | Forward as remote `podman compose down` and mark the deployment as removed |
| `ps` / `logs` / `restart` / `exec` / `build` | `podrun <cmd> [args]` | Forward as `podman compose <cmd> [args]` in the remote project folder |
| `domain` / `deploy` / `export` / `info` / `clone` | — | **[Unfinished]** See Implementation Status below |

Any other command returns `unsupported command`.

### Implementation Status

| Item | Status | Planned Scope |
|---|---|---|
| Podman Compose deploy (`up` / `down` / `clear` / passthrough commands) | Done | — |
| Deployment record API (`/api/pod/*`) | Done | — |
| k3s runtime (`--type=k3s`) | **[Unfinished]** | Deploy to k3s with the same command; currently only stored in the `target` field |
| `deploy` | **[Unfinished]** | Deploy the project to Kubernetes |
| `export` | **[Unfinished]** | Export the project as a Pod manifest |
| `domain` | **[Unfinished]** | Assign a domain to a Pod (`domains` table already reserved) |
| `info` | **[Unfinished]** | Show project info |
| `clone` | **[Unfinished]** | Clone the remote project back to local |

### Flags

| Flag | Description |
|---|---|
| `-d`, `--detach` | Run in the background; also forwarded to `podman compose` |
| `--folder=<path>`, `--folder <path>` | Local project folder; defaults to the current directory |
| `./<dir>`, `/<dir>` | Positional local project folder (`--folder` takes precedence) |
| `-f <file>` | Compose file; means follow under `logs`. Multiple `-f` flags are not supported |
| `-u <uid>`, `-u=<uid>` | Override the deployment UID |
| `--type=<target>`, `--type <target>` | **[Unfinished]** Runtime target; defaults to `podman`. `k3s` is planned: it is currently only stored in the deployment's `target` field and still runs through Podman Compose |

### API Endpoints

The API server listens on `:8080` and responds with the string `ok` or `{"data": [...]}`.

| Method | Path | Body | Description |
|---|---|---|---|
| `GET` | `/api/health` | — | Health check; returns `ok` |
| `GET` | `/api/pod/list` | — | List deployments with `dismiss = 0` |
| `POST` | `/api/pod/upsert` | `Pod` | Insert or update a deployment by `uid` and reset `dismiss = 0` |
| `POST` | `/api/pod/update/:uid` | `Pod` (uses `status`, `dismiss`) | Update the status and removal flag of a deployment |
| `POST` | `/api/pod/record/insert` | `Record` | Insert an operation record by `uid` |

### Tables

| Table | Purpose | Key Columns |
|---|---|---|
| `pods` | One row per deployment; `uid` is unique | `uid`, `pod_uid`, `pod_name`, `local_dir`, `remote_dir`, `file`, `target`, `status`, `hostname`, `ip`, `replicas`, `dismiss` |
| `records` | Operation log linked to `pods.id` | `content` (`up` / `sync` / `overwrite` / `down` / `clear`...), `hostname`, `ip` |
| `domains` | Reserved for the `domain` command (**[Unfinished]**) | `container_name`, `domain` |

### `Pod` JSON Fields

| Field | Type | Description |
|---|---|---|
| `uid` | `string` | `md5("<MAC>@<local-absolute-path>")`; falls back to the hostname when no MAC is found |
| `pod_id` | `string` | Podman Pod ID; the remote folder name when the lookup fails |
| `pod_name` | `string` | Podman Pod name; the remote folder name when the lookup fails |
| `local_dir` | `string` | Local project absolute path |
| `remote_dir` | `string` | `/home/podrun/<folder-name>_<first-8-hash-chars>` |
| `file` | `string` | Compose file passed via `-f` |
| `target` | `string` | Value of `--type`; **[Unfinished]** k3s is not implemented yet and is recorded only |
| `status` | `string` | Deployment status; the CLI writes `starting` |
| `hostname` | `string` | Hostname of the machine running the CLI |
| `ip` | `string` | First non-loopback IPv4 of the machine running the CLI |
| `replicas` | `int` | Replica count; always `1` |
| `dismiss` | `int` | `0` active, `1` removed |

***

©️ 2025 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
