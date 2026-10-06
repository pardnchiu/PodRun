# PodRun - 技術文件

最後更新：2026-10-06

> 返回 [README](./README.zh.md)

## 前置需求

| 需求 | 位置 | 說明 |
|---|---|---|
| Go 1.25.1+ | 編譯端 | 版本取自 `go.mod` |
| C 編譯器（`CGO_ENABLED=1`） | 編譯端（API Server） | `github.com/mattn/go-sqlite3` 需要 cgo |
| `sshpass`、`rsync`、`ssh`、`curl`、`unzip` | 本地（CLI） | 缺少時啟動即自動安裝：macOS 走 `brew`，Linux 依序嘗試 `apt`／`dnf`／`yum`／`pacman` |
| macOS 或 Linux | 本地（CLI） | 其他作業系統在缺套件時直接中止 |
| SSH 密碼登入 | 遠端主機 | CLI 以 `sshpass` 帶密碼連線 |
| Podman + `podman compose` | 遠端主機 | 容器執行環境 |
| `/home/podrun` 寫入權限 | 遠端主機 | 遠端專案目錄固定建立於此路徑下 |

## 安裝

### 從原始碼編譯

```bash
git clone https://github.com/pardnchiu/PodRun.git
cd PodRun
go build -o podrun ./cmd/cli
CGO_ENABLED=1 go build -o podrun-api ./cmd/api
```

### 使用 go install（僅 CLI）

```bash
go install github.com/pardnchiu/PodRun/cmd/cli@latest
```

`go install` 產出的執行檔名稱為 `cli`（取自套件目錄名），可自行改名為 `podrun`。

## 設定

### 環境變數

CLI 與 API Server 啟動時都會以 `godotenv` 讀取**當前工作目錄**的 `.env`；讀不到只會記錄 warning，仍以系統環境變數為準。

| 變數 | 使用端 | 必填 | 預設值 | 說明 |
|---|---|---|---|---|
| `PODRUN_SERVER` | CLI | 是 | — | 遠端主機 Hostname 或 IP |
| `PODRUN_USERNAME` | CLI | 是 | — | 遠端 SSH 使用者 |
| `PODRUN_PASSWORD` | CLI | 是 | — | 遠端 SSH 密碼（交給 `sshpass`） |
| `DB_PATH` | API Server | 否 | 容器內（存在 `/.dockerenv`）：`/data/database.db`；其他：`~/.podrun/database.db` | SQLite 檔案路徑；使用預設路徑時自動建立 `~/.podrun/` |
| `ALLOW_EMAILS` | API Server | 否 | — | 列於 `.env.expample`，目前程式未啟用 |

### 設定檔

```bash
cp .env.expample .env
```

```dotenv
PODRUN_SERVER=192.168.1.100
PODRUN_USERNAME=podrun
PODRUN_PASSWORD=yourpassword
```

## 使用方式

### 啟動 API Server

API Server 啟動時以相對路徑讀取 `sql/create.sql` 建表，須在 repo 根目錄執行；CLI 固定連往 `localhost:8080`，兩者需在同一台機器。

```bash
cd PodRun
./podrun-api
```

### 基本：部署當前專案

在含有 `docker-compose.yml` 或 `docker-compose.yaml` 的專案目錄執行：

```bash
cd ~/projects/my-app
podrun up -d
```

`up` 依序執行：

1. 以 SSH 測試連線，失敗即中止
2. 在遠端建立 `/home/podrun/<目錄名>_<雜湊前 8 碼>/`
3. 遠端目錄非空時，以 rsync dry-run 列出差異，有變更需輸入 `y` 確認；接著以 `rsync -avz --delete` 同步（排除 `node_modules/`、`vendor/`、`.git/`、`.venv/`、`.next/`、`*.log` 等）
4. 複製 compose 檔為 `docker-compose.podrun.yml`，移除主機 Port 綁定（`8080:80` → `80`），並為 `./` 開頭的 Volume 補上 `:z`
5. 先以 `down -v` 清掉舊容器，再執行 `podman compose -f docker-compose.podrun.yml up -d`
6. 查詢 Pod ID／名稱，`-d` 模式下印出各容器 Port 對應
7. 將部署資訊與操作紀錄送往 API Server

未加 `-d` 時，`up` 在前景執行；按下 `Ctrl+C` 會在遠端觸發 `podman compose down`。

### 進階：指定專案路徑與 compose 檔

```bash
# 指定本地專案目錄
podrun up -d --folder=/path/to/project

# 直接帶目錄路徑（須以 ./ 或 / 開頭且存在）
podrun up -d ./services/api

# 指定 compose 檔；未指定 --folder 時以該檔所在目錄為專案目錄
podrun up -d -f ./services/api/docker-compose.yml

# 指定部署 UID（覆寫自動產生的雜湊）
podrun up -d -u my-app-prod
```

### 進階：維運指令

```bash
# 查看容器狀態
podrun ps

# 持續追蹤日誌（logs 後的 -f 視為 follow 而非檔案）
podrun logs -f

# 進入容器
podrun exec web sh

# 停止並移除容器
podrun down

# 完整清除：容器、Volume、映像與遠端專案目錄
podrun clear
```

`clear` 以 `podman run --privileged alpine` 刪除遠端目錄，可清掉 rootless 容器產生、一般使用者無權刪除的檔案。

## 命令列參考

### 指令

| 指令 | 語法 | 說明 |
|---|---|---|
| `up` | `podrun up [-d] [flags]` | 同步、改寫 compose、重建並啟動容器、寫入部署紀錄 |
| `clear` | `podrun clear` | `down -v`、`down --rmi all`、刪除遠端專案目錄 |
| `down` | `podrun down [args]` | 轉送為遠端 `podman compose down`，並將部署標記為已移除 |
| `ps`／`logs`／`restart`／`exec`／`build` | `podrun <cmd> [args]` | 在遠端專案目錄轉送為 `podman compose <cmd> [args]` |
| `domain`／`deploy`／`export`／`info`／`clone` | — | **[未完成]** 見下方「實作狀態」 |

其餘指令回傳 `unsupported command`。

### 實作狀態

| 項目 | 狀態 | 規劃內容 |
|---|---|---|
| Podman Compose 部署（`up`／`down`／`clear`／轉送指令） | 已完成 | — |
| 部署紀錄 API（`/api/pod/*`） | 已完成 | — |
| k3s Runtime（`--type=k3s`） | **[未完成]** | 以同一指令部署至 k3s；目前僅寫入 `target` 欄位 |
| `deploy` | **[未完成]** | 將專案部署至 Kubernetes |
| `export` | **[未完成]** | 將專案匯出為 Pod Manifest |
| `domain` | **[未完成]** | 為 Pod 設定網域（已預留 `domains` 資料表） |
| `info` | **[未完成]** | 顯示專案資訊 |
| `clone` | **[未完成]** | 將遠端專案複製回本地 |

### 旗標

| 旗標 | 說明 |
|---|---|
| `-d`、`--detach` | 背景執行；同時轉送給 `podman compose` |
| `--folder=<path>`、`--folder <path>` | 本地專案目錄，預設為當前目錄 |
| `./<dir>`、`/<dir>` | 以位置參數指定本地專案目錄（`--folder` 優先） |
| `-f <file>` | 指定 compose 檔；`logs` 指令下改為 follow。不支援多個 `-f` |
| `-u <uid>`、`-u=<uid>` | 覆寫部署 UID |
| `--type=<target>`、`--type <target>` | **[未完成]** Runtime 目標，預設 `podman`；`k3s` 為後續實作項目，目前僅寫入部署紀錄的 `target` 欄位，實際仍以 Podman Compose 執行 |

### API 端點

API Server 監聽 `:8080`，回應 `ok` 字串或 `{"data": [...]}`。

| 方法 | 路徑 | Body | 說明 |
|---|---|---|---|
| `GET` | `/api/health` | — | 健康檢查，回傳 `ok` |
| `GET` | `/api/pod/list` | — | 列出 `dismiss = 0` 的部署 |
| `POST` | `/api/pod/upsert` | `Pod` | 依 `uid` 新增或更新部署，並重設 `dismiss = 0` |
| `POST` | `/api/pod/update/:uid` | `Pod`（取 `status`、`dismiss`） | 更新指定部署的狀態與移除旗標 |
| `POST` | `/api/pod/record/insert` | `Record` | 依 `uid` 寫入一筆操作紀錄 |

### 資料表

| 資料表 | 用途 | 主要欄位 |
|---|---|---|
| `pods` | 每個部署一列，`uid` 唯一 | `uid`、`pod_uid`、`pod_name`、`local_dir`、`remote_dir`、`file`、`target`、`status`、`hostname`、`ip`、`replicas`、`dismiss` |
| `records` | 操作紀錄，關聯 `pods.id` | `content`（`up`／`sync`／`overwrite`／`down`／`clear`...）、`hostname`、`ip` |
| `domains` | 預留給 `domain` 指令（**[未完成]**） | `container_name`、`domain` |

### `Pod` JSON 欄位

| 欄位 | 型別 | 說明 |
|---|---|---|
| `uid` | `string` | `md5("<MAC>@<本地絕對路徑>")`，無法取得 MAC 時改用 Hostname |
| `pod_id` | `string` | Podman Pod ID；查不到時為遠端目錄名 |
| `pod_name` | `string` | Podman Pod 名稱；查不到時為遠端目錄名 |
| `local_dir` | `string` | 本地專案絕對路徑 |
| `remote_dir` | `string` | `/home/podrun/<目錄名>_<雜湊前 8 碼>` |
| `file` | `string` | `-f` 指定的 compose 檔 |
| `target` | `string` | `--type` 的值；**[未完成]** k3s 尚未實作，僅作紀錄 |
| `status` | `string` | 部署狀態，CLI 寫入 `starting` |
| `hostname` | `string` | 執行 CLI 的主機名稱 |
| `ip` | `string` | 執行 CLI 的主機第一個非 loopback IPv4 |
| `replicas` | `int` | 副本數，固定為 `1` |
| `dismiss` | `int` | `0` 有效、`1` 已移除 |

***

©️ 2025 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
