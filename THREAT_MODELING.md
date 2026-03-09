# Sean AIgent 威脅建模分析 (Threat Modeling)

本文件由安全工程師視角出發，針對 **Sean AIgent v2** 預約系統進行安全評估。我們採用微軟開發的 **STRIDE 模型** 來識別潛在威脅，並針對現有架構提出防禦建議。

## 1. 系統架構概覽 (Architecture Overview)
- **後端**: Go 1.25 (Gin Framework)
- **資料庫**: MongoDB (具有 EventStore 事件溯源架構)
- **前端**: LINE LIFF (HTMX + Alpine.js)
- **基礎設施**: Docker + Dagger CI/CD

---

## 2. STRIDE 威脅分析

| 威脅類別 | 說明 (Threat Description) | 風險點 (Potential Vulnerabilities) | 建議緩解措施 (Mitigation Strategies) |
| :--- | :--- | :--- | :--- |
| **S - 仿冒 (Spoofing)** | 攻擊者偽造 LINE User ID 或盜用 `user_token` | `liff-init.js` 中的 Token 傳遞邏輯 | 1. 後端強制執行 LINE Access Token 即時驗證。<br>2. 嚴禁僅依賴 URL 參數中的 UserID 進行敏感操作。 |
| **T - 竄改 (Tampering)** | 惡意修改 `bookingId` 或 `status` | 批次更新 API `/v2/admin/checkin/batch-update` | 1. 實作「水平權限校驗」，確保操作者對該對象擁有所有權。<br>2. 維持現有 CSRF 保護機制。 |
| **R - 抵賴 (Repudiation)** | 使用者或教練否認其操作行為 | 關鍵業務邏輯（點名、請假、取消） | 1. **(強項)** 善用 `EventStore` 記錄完整操作日誌。<br>2. 事件 Payload 需包含操作者的 `ActorID` 與 `IP`。 |
| **I - 資訊洩露 (Info Disclosure)** | 敏感學員 PII 數據（電話、姓名）外洩 | 統計報表注入前端的 JSON 資料 (`data-stats`) | 1. 精簡前端 VO (Value Object) 欄位，僅傳輸必要顯示的資訊。<br>2. 對 CSV 匯出 API 實施嚴格的導出頻率限制。 |
| **D - 阻斷服務 (DoS)** | 惡意腳本瞬間佔滿名額或耗盡伺服器資源 | 預約 API 與高負載的聚合查詢 (Aggregation) | 1. 引入 **Rate Limiting (限流)**，限制單一用戶請求頻率。<br>2. 針對高負載查詢實作快取與分頁保護。 |
| **E - 特權提升 (Privilege Elevation)** | 一般用戶存取管理端 `/v2/admin/*` 路徑 | Middleware 中的 Admin Role 判斷邏輯 | 1. 採用「預設拒絕 (Default Deny)」策略。<br>2. 敏感管理介面實作二要素驗證 (2FA) 或限制存取來源 IP。 |

---

## 3. 針對 v2.1 / v2.2 新功能的專屬風險評估

### A. 帳務管理 (Accounting & Payments)
*   **威脅**: 收款狀態被前端攔截竄改。
*   **防護**: 
    - 收款紀錄應作為不可變事件 (Immutable Event) 寫入。
    - 財務對帳邏輯應封閉於後端 Domain 層，不應在前端進行金額計算。

### B. 團隊與校隊管理 (Team Management)
*   **威脅**: 跨隊資料存取 (A 隊教練讀取 B 隊學員資料)。
*   **防護**: 
    - 實作基於 **TeamID** 的資料隔離 (Data Partitioning)。
    - 確保所有 Repository 查詢都包含 `team_id` 過濾條件。

---

## 4. 安全工程師建議清單 (Action Items)

### 🔴 高優先級 (Immediate Action)
1.  **水平權限校驗**: 檢查所有 `Update` 類型的 API，確保 User A 不能修改 User B 的資料。
2.  **API 限流**: 在 Gin 中加入全局與特定路由的限流 Middleware。

### 🟡 中優先級 (Scheduled Improvement)
1.  **PII 脫敏**: 在管理介面顯示家長電話時，應考慮部分隱藏 (e.g., 0912****56)。
2.  **安全性標頭**: 確保 Nginx 或 Go 提供正確的 `Content-Security-Policy` (CSP) 與 `Strict-Transport-Security` (HSTS)。

### 🟢 低優先級 (Best Practices)
1.  **Go 1.25 安全更新**: 持續追蹤 Go 標準庫對於 TLS 與加密演算法的更新。
2.  **事件審計**: 定期稽核 `EventStore` 中的異常變更行為。

---
*文件更新日期: 2026-03-08 | 評估人員: Gemini CLI Security Engineer*
