# Sean AIgent 開發進度追蹤 (TODO)

## 🏗️ 系統架構：事件驅動非同步流 (Event-Driven Stream)
- [x] 定義 `Event` 介面與基本事件結構 (`TypedEvent`)
- [x] 實作內部的 `EventBus` (基於 Go Channels & Registry 模式，含進度追趕與 Recover 機制)
- [x] 實作事件持久化儲存 `EventStore` (MongoDB)
- [x] 重構現有 UseCase，改為發送 `AppointmentStatusChanged` 等領域事件
- [x] 將 `CacheWorker` 轉型為事件訂閱者 (Subscriber)

## 📊 經營分析與數據報表優化 (預聚合快照表方案)

### Phase 1: 基礎建設 & 資料庫設計
- [x] 定義 `UserMonthlyStat` Domain Entity 與 JSON 結構
- [x] 實作 `StatsRepository` (MongoDB)
    - [x] `UpsertUserMonthlyStats`: 更新或新增單一用戶月統計
    - [x] `FindMonthlyStats`: 支援年份、月份、分頁、搜尋
    - [x] 建立資料庫索引: `{ year: 1, month: 1, user_id: 1 }`
- [x] Wire 依賴注入設定

### Phase 2: 事件訂閱者與計算邏輯
- [ ] 實作 `UserMonthlyStatsSubscriber`
    - [ ] 監聽 `AppointmentStatusChanged` 事件
    - [ ] 核心運算: 執行單一用戶的 Aggregation Pipe
    - [ ] 資料持久化: 呼叫 Repo 寫入 `user_monthly_stats` 集合
- [ ] 實作 `BatchSyncMonthlyStatsUseCase` (供手動或 Cron 全量校準使用)

### Phase 3: 狀態連動 (透過事件)
- [ ] 在 `AdminBatchUpdateAttendance` 成功後發送事件
- [ ] 在 `AdminCreateWalkIn` (現場加人) 成功後發送事件
- [ ] 在 `AutoMarkAbsent` (Cron) 成功後發送事件

### Phase 4: Cron 自動化校準
- [ ] 註冊 `POST /cron/sync-all-stats` API 節點
- [ ] 在 `cmd/cron.go` 設定定期排程 (建議每天 03:00)
- [ ] 實作校準邏輯: 掃描過去 30 天內有變動的用戶並重新聚合

### Phase 5: 報表功能對接
- [ ] 重構 `getUserReport` (學員月報表): 改從快照表讀取 + Server-side 分頁
- [ ] 重構 `getAnalytics` (經營分析看板): 使用快照表計算趨勢
- [ ] 實作 CSV 匯出功能 (基於快照表)

---

## ✅ 已完成項目
- [x] **事件系統效能優化**：支援自定義 `Marshaler` / `Unmarshaler`，實作 `sync.Once` 延遲序列化，大幅降低記憶體與 CPU 消耗。
- [x] V2 管理看板、批次提交點名、自動缺席判定 Cron 等
