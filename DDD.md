smart-booking-system/
├── cmd/
│   └── server/             # 程式進入點 (main.go)
├── internal/               # 存放主要業務邏輯，外部無法存取
│   ├── booking/            # 核心限界上下文 (Core Context)
│   │   ├── domain/         # 1. 領域層：實體、狀態機、業務規則
│   │   │   ├── appointment.go
│   │   │   ├── leave_record.go
│   │   │   ├── slot.go
│   │   │   └── repository.go (Interface)
│   │   ├── app/            # 2. 應用層：Use Cases (編排業務邏輯)
│   │   │   ├── booking_service.go
│   │   │   ├── attendance_service.go
│   │   │   └── leave_service.go
│   │   └── infra/          # 3. 基礎設施層：實作外部介面
│   │       ├── mysql_repo/ # 數據庫實作 (GORM/SQL)
│   │       └── leave_db/   # 獨立的請假紀錄 DB 實作
│   ├── interaction/        # 外部互動上下文 (LINE/LLM)
│   │   ├── line/           # LINE Bot 適配器
│   │   │   ├── bot_handler.go
│   │   │   ├── flex_message.go (存放 Flex Message JSON)
│   │   │   └── liff_handler.go (處理網頁預約)
│   │   └── ai/             # LLM 整合 (OpenAI/Gemini)
│   │       └── parser.go   # 解析排課、請假原因分類
│   └── shared/             # 共享模組 (如：通用錯誤代碼、工具)
├── web/                    # 前端網頁專案 (Liff/Admin Dashboard)
│   ├── parent-booking/     # 家長選課網頁
│   └── coach-attendance/   # 教練點名網頁
├── api/                    # API 定義文件 (Swagger/OpenAPI)
├── configs/                # 設定檔 (YAML/Env)
├── deployments/            # 部署相關 (Docker, k8s)
├── go.mod
└── README.md               # 剛才彙整的系統定義文件