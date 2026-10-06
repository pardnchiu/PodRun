# PodRun - 架構

最後更新：2026-10-06

> 返回 [README](./README.zh.md)

## 概覽

```mermaid
graph LR
    subgraph 本地主機
        CLI[podrun CLI<br/>cmd/cli]
        API[API Server<br/>cmd/api :8080]
        DB[(SQLite)]
    end
    subgraph 遠端主機
        Dir[/home/podrun/名稱_雜湊/]
        Compose[Podman Compose]
    end
    CLI -->|rsync over sshpass| Dir
    CLI -->|ssh 指令| Compose
    Compose --> Dir
    CLI -.->|未完成| K3s[k3s]
    CLI -->|HTTP POST| API
    API --> DB
```

## Module: CLI 入口（cmd/cli）

依序檢查依賴套件、環境變數、解析參數、測試 SSH，再分派指令。

```mermaid
graph TB
    subgraph cmd/cli
        Init[godotenv 載入 .env] --> Rely[CheckRelyPackages<br/>缺套件自動安裝]
        Rely --> Env[CheckENV<br/>PODRUN_SERVER／USERNAME／PASSWORD]
        Env --> New[command.New]
        New --> Test[SSHTest<br/>ConnectTimeout=3]
        Test --> Switch{RemoteArgs 首項}
        Switch -->|domain／deploy| Noop[無動作]
        Switch -->|其他| Compose[ComposeCMD]
    end
    Brew[brew／apt／dnf／yum／pacman] -.-> Rely
```

## Module: command（internal/command）

參數解析、本地／遠端路徑推導與 compose 指令執行。

```mermaid
graph TB
    subgraph internal/command
        Parse[parseArgs<br/>-d／-f／-u／--folder／--type] --> Local[getLocalDir<br/>需含 docker-compose.yml/yaml]
        Local --> Remote[setRemoteDir<br/>md5 MAC@路徑]
        Remote --> Arg[PodmanArg]
        Arg --> Dispatch{ComposeCMD}
        Dispatch -->|up| Up[up]
        Dispatch -->|clear| Clear[clear]
        Dispatch -->|down／ps／logs／restart／exec／build| Run[runCMD 轉送]
        Up --> Rsync[RsyncToRemote<br/>dry-run 預覽 + 確認]
        Up --> Modify[ModifyComposeFile<br/>移除 Port、補 :z]
        Up --> Report[upsertPod／recordPod]
        Clear --> Report
        Run --> Report
    end
    Rsync --> Utils[internal/utils]
    Modify --> Utils
    Report --> HTTP[localhost:8080]
```

## Module: utils（internal/utils）

封裝 `sshpass` 系列指令與本機資訊取得。

```mermaid
graph TB
    subgraph internal/utils
        CheckENV --> Podrun[Podrun<br/>Remote = user@server]
        Podrun --> SSHTest
        Podrun --> SSHRun[SSHRun<br/>ssh -tt，綁定 stdio]
        Podrun --> SSEOutput[SSEOutput<br/>擷取輸出]
        CMDRun[CMDRun]
        CMDOutput[CMDOutput]
        GetMAC
        GetHostName
        GetLocalIP[GetLocalIP<br/>首個非 loopback IPv4]
    end
    SSHRun --> Sshpass[sshpass + ssh]
    SSEOutput --> CMDOutput
    CMDOutput --> Sshpass
```

## Module: API Server（cmd/api + internal/handler + internal/database）

Gin 路由搭配 SQLite 儲存部署與操作紀錄。

```mermaid
graph TB
    subgraph cmd/api
        Path{DB_PATH?} -->|有| Open
        Path -->|無，/.dockerenv 存在| Docker[/data/database.db] --> Open
        Path -->|無| Home[~/.podrun/database.db] --> Open
        Open[NewSQLite<br/>執行 sql/create.sql]
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
    ListPods --> DB[(pods／records／domains)]
    UpsertPod --> DB
    UpdatePod --> DB
    InsertRecord --> DB
```

## 資料流

`podrun up -d` 的完整流程：

```mermaid
sequenceDiagram
    participant U as 使用者
    participant C as podrun CLI
    participant R as 遠端主機
    participant A as API Server
    U->>C: podrun up -d
    C->>R: ssh exit（連線測試）
    C->>R: mkdir -p /home/podrun/名稱_雜湊
    C->>R: 檢查遠端目錄是否為空
    alt 非空
        C->>R: rsync -avni --delete（dry-run）
        C->>U: 顯示差異並詢問 y/N
        U-->>C: y
        C->>A: record overwrite
    else 空
        C->>A: record sync
    end
    C->>R: rsync -avz --delete
    C->>R: cp 為 docker-compose.podrun.yml + sed／awk 改寫
    C->>R: podman compose down -v
    C->>A: update dismiss=1
    C->>R: podman compose -f docker-compose.podrun.yml up -d
    C->>R: podman pod ps 取得 Pod ID／名稱
    C->>R: podman ps 取得 Port 對應
    C->>A: upsert pod（dismiss=0）
    C->>A: record up
```

## 狀態機

`pods.dismiss` 的轉換：

```mermaid
stateDiagram-v2
    [*] --> 有效: up（upsert）
    有效 --> 已移除: down／clear／重新 up 前清理
    已移除 --> 有效: up（upsert 重設 dismiss=0）
    已移除 --> [*]
```

***

©️ 2025 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
