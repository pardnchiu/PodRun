# PodRun - Architecture

Last updated: 2026-10-06

> Back to [README](../README.md)

## Overview

```mermaid
graph LR
    subgraph Local Host
        CLI[podrun CLI<br/>cmd/cli]
        API[API Server<br/>cmd/api :8080]
        DB[(SQLite)]
    end
    subgraph Remote Host
        Dir[/home/podrun/name_hash/]
        Compose[Podman Compose]
    end
    CLI -->|rsync over sshpass| Dir
    CLI -->|ssh commands| Compose
    Compose --> Dir
    CLI -.->|unfinished| K3s[k3s]
    CLI -->|HTTP POST| API
    API --> DB
```

## Module: CLI Entry (cmd/cli)

Checks dependencies, environment variables, parses arguments, tests SSH, then dispatches the command.

```mermaid
graph TB
    subgraph cmd/cli
        Init[godotenv loads .env] --> Rely[CheckRelyPackages<br/>auto-install missing packages]
        Rely --> Env[CheckENV<br/>PODRUN_SERVER / USERNAME / PASSWORD]
        Env --> New[command.New]
        New --> Test[SSHTest<br/>ConnectTimeout=3]
        Test --> Switch{First RemoteArg}
        Switch -->|domain / deploy| Noop[No action]
        Switch -->|other| Compose[ComposeCMD]
    end
    Brew[brew / apt / dnf / yum / pacman] -.-> Rely
```

## Module: command (internal/command)

Argument parsing, local/remote path derivation, and compose command execution.

```mermaid
graph TB
    subgraph internal/command
        Parse[parseArgs<br/>-d / -f / -u / --folder / --type] --> Local[getLocalDir<br/>requires docker-compose.yml/yaml]
        Local --> Remote[setRemoteDir<br/>md5 MAC@path]
        Remote --> Arg[PodmanArg]
        Arg --> Dispatch{ComposeCMD}
        Dispatch -->|up| Up[up]
        Dispatch -->|clear| Clear[clear]
        Dispatch -->|down / ps / logs / restart / exec / build| Run[runCMD passthrough]
        Up --> Rsync[RsyncToRemote<br/>dry-run preview + confirm]
        Up --> Modify[ModifyComposeFile<br/>strip ports, add :z]
        Up --> Report[upsertPod / recordPod]
        Clear --> Report
        Run --> Report
    end
    Rsync --> Utils[internal/utils]
    Modify --> Utils
    Report --> HTTP[localhost:8080]
```

## Module: utils (internal/utils)

Wraps `sshpass`-based commands and local host information lookups.

```mermaid
graph TB
    subgraph internal/utils
        CheckENV --> Podrun[Podrun<br/>Remote = user@server]
        Podrun --> SSHTest
        Podrun --> SSHRun[SSHRun<br/>ssh -tt, bound stdio]
        Podrun --> SSEOutput[SSEOutput<br/>captured output]
        CMDRun[CMDRun]
        CMDOutput[CMDOutput]
        GetMAC
        GetHostName
        GetLocalIP[GetLocalIP<br/>first non-loopback IPv4]
    end
    SSHRun --> Sshpass[sshpass + ssh]
    SSEOutput --> CMDOutput
    CMDOutput --> Sshpass
```

## Module: API Server (cmd/api + internal/handler + internal/database)

Gin routes backed by SQLite store deployments and operation records.

```mermaid
graph TB
    subgraph cmd/api
        Path{DB_PATH?} -->|set| Open
        Path -->|unset, /.dockerenv exists| Docker[/data/database.db] --> Open
        Path -->|unset| Home[~/.podrun/database.db] --> Open
        Open[NewSQLite<br/>runs sql/create.sql]
    end
    subgraph internal/handler
        Routes[NewRoutes :8080]
        Routes --> List[GET /api/pod/list]
        Routes --> Upsert[POST /api/pod/upsert]
        Routes --> Update[POST /api/pod/update/:uid]
        Routes --> Insert[POST /api/pod/record/insert]
        Routes --> Health[GET /api/health]
    end
    subgraph internal/database
        ListPods
        UpsertPod
        UpdatePod
        InsertRecord
    end
    Open --> Routes
    List --> ListPods
    Upsert --> UpsertPod
    Update --> UpdatePod
    Insert --> InsertRecord
    ListPods --> DB[(pods / records / domains)]
    UpsertPod --> DB
    UpdatePod --> DB
    InsertRecord --> DB
```

## Data Flow

Full flow of `podrun up -d`:

```mermaid
sequenceDiagram
    participant U as User
    participant C as podrun CLI
    participant R as Remote Host
    participant A as API Server
    U->>C: podrun up -d
    C->>R: ssh exit (connection test)
    C->>R: mkdir -p /home/podrun/name_hash
    C->>R: check whether the remote folder is empty
    alt not empty
        C->>R: rsync -avni --delete (dry-run)
        C->>U: show diff and ask y/N
        U-->>C: y
        C->>A: record overwrite
    else empty
        C->>A: record sync
    end
    C->>R: rsync -avz --delete
    C->>R: cp to docker-compose.podrun.yml + sed / awk rewrite
    C->>R: podman compose down -v
    C->>A: update dismiss=1
    C->>R: podman compose -f docker-compose.podrun.yml up -d
    C->>R: podman pod ps for Pod ID / name
    C->>R: podman ps for port mappings
    C->>A: upsert pod (dismiss=0)
    C->>A: record up
```

## State Machine

Transitions of `pods.dismiss`:

```mermaid
stateDiagram-v2
    [*] --> Active: up (upsert)
    Active --> Removed: down / clear / cleanup before re-up
    Removed --> Active: up (upsert resets dismiss=0)
    Removed --> [*]
```

***

©️ 2025 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
