# SeanAIgent

SeanAIgent 是一個整合了 LINE Bot 的智慧預約管理系統，主要用於訓練課程的預約與報到。它同時提供了一個網頁介面，方便管理者進行操作與設定。

## ✨ 功能

- **課程管理**：建立、查詢、管理訓練課程與日期。
- **學員預約**：使用者可以透過 LINE Bot 查詢可預約的時段並完成預約。
- **學員報到**：使用者可以透過 LINE Bot 進行上課前的報到。
- **網頁儀表板**：提供管理者使用的網頁介面，用於管理課程、檢視預約狀態等。
- **AI 整合**：利用大型語言模型（LLM）提供更智慧的互動與功能。
- **LLM 驅動的訓練時段建立**：教練可以透過自然語言指令，讓 LLM 自動建立訓練課程時段，例如：「請在下週三下午兩點到四點，新增一個容納 10 人的訓練課程。」

## 🛠️ 技術棧

- **後端 (Go)**: 專案主要的後端語言，提供高效能的服務。
- **網頁框架 (Gin)**: 一個輕量級的 Go 網頁框架，用於處理 HTTP 請求和路由。
- **資料庫 (MongoDB)**: 使用 NoSQL 資料庫來儲存課程、預約等相關資料。
- **前端樣板 ([templ](https://github.com/a-h/templ))**: 一個 Go 的樣板引擎，用於將 Go 程式碼直接編譯成 HTML，實現元件化的前端開發。
- **CSS 框架 (Tailwind CSS)**: 一個 Utility-First 的 CSS 框架，用於快速建構現代化的使用者介面。
- **命令列介面 (Cobra)**: 用於建立強大的 CLI 應用程式，是本專案 `serve`, `mcp` 等指令的基礎。
- **大型語言模型整合 (LangChainGo)**: 透過 LangChainGo 框架，整合大型語言模型 (LLM) 來實現自然語言處理功能。
- **即時通訊 (LINE Bot SDK)**: 用於與 LINE Platform 對接，實現 LINE Bot 的訊息收發與互動。
- **容器化 (Docker & Docker Compose)**: 用於打包應用程式及其依賴，並在不同環境中提供一致的部署與執行體驗。

## 🚀 如何開始

### 環境準備

在開始之前，請確保您已經安裝了以下工具：

- [Go](https://golang.org/) (版本 1.24+)
- [Make](https://www.gnu.org/software/make/)
- [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)
- [templ CLI](https://templ.guide/getting-started/installation)
- [Tailwind CSS CLI](https://tailwindcss.com/docs/installation)
- [Air](https://github.com/cosmtrek/air) (用於 Go 的 Hot-Reload)

### 設定

本專案的設定檔是透過 volume 掛載進容器的。請在您的主目錄下建立設定檔：

1.  建立資料夾：
    ```bash
    mkdir -p ~/.Secret/seanAigent/
    ```

2.  建立設定檔 `~/.Secret/seanAigent/config.yaml`。根據專案的程式碼，設定檔可能需要包含以下內容（請根據您的實際環境修改）：

    ```yaml
    # 範例 config.yaml
    mongodb:
      uri: "mongodb://user:password@host:port"
      database: "your_db_name"

    line:
      channel_secret: "YOUR_LINE_CHANNEL_SECRET"
      channel_token: "YOUR_LINE_CHANNEL_TOKEN"

    llm:
      # LLM 相關設定
      api_key: "YOUR_LLM_API_KEY"
    ```

### 開發模式

開發模式會啟動 hot-reload，讓您在修改程式碼後能立即看到效果。

```bash
make dev
```

這個指令會同時執行三個任務：
1.  `make tailwind`: 監控 `assets/css/input.css` 的變化並產生 `output.css`。
2.  `make templ`: 監控 `.templ` 檔案的變化並產生對應的 Go 程式碼。
3.  `make server`: 使用 `air` 監控 Go 程式碼的變化，並在變動時自動重新編譯和執行。

服務將會啟動在 `http://localhost:8082`。

### 正式環境 (使用 Docker)

您可以使用 Docker Compose 來啟動正式環境的服務。

```bash
docker-compose up -d
```

這個指令會根據 `docker-compose.yml` 的設定，在背景啟動 `mcp` 和 `console` 兩個服務。請確保您的 Traefik 或其他反向代理設定正確，以便能透過 `seanaigent.94peter.dev` 存取網頁介面。

## 🏗️ 建置

您可以使用 `make` 來建置 Docker image。

- **建置單一平台的 image**:
  ```bash
  make build
  ```
  這會建置一個名為 `seanaigent:latest` 的 image。

- **建置跨平台的 image**:
  ```bash
  make multi-build
  ```
  這會建置支援 `linux/amd64` 和 `linux/arm64` 的 image，並將其推送到 Docker Hub 上的 `94peter/seanaigent:latest`。

## 📂 專案結構

```
.
├── cmd/            # Cobra CLI 指令
├── components/     # templ UI 元件
├── internal/       # 專案內部的主要商業邏輯
│   ├── dao/
│   ├── db/
│   ├── handler/
│   ├── mcp/
│   └── service/
├── pkg/            # 可供外部專案使用的共享套件
├── templates/      # 頁面層級的 templ 樣板
├── assets/         # CSS, JavaScript 等前端資源
├── main.go         # 程式進入點
├── Makefile        # 自動化指令
├── go.mod          # Go 模組與依賴
└── docker-compose.yml # Docker Compose 設定
```

## 📄 授權

本專案採用 [MIT License](LICENSE) 授權。
