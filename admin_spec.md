# 📋 管理者後台 (Admin Dashboard) 規格書

## 一、 總覽
本後台專為教練與管理人員設計，旨在提供即時的場次狀態監控、精確的學員出席數據分析，以及便捷的現場管理工具。系統採用 **Dark Mode** 視覺風格，確保在各種環境下（包含戶外強光與室內）皆具備高度可讀性。

---

## 二、 核心功能頁面

### 1. 場次監控看板 (Session Monitor)
*   **路徑**: `/admin/dashboard`
*   **功能**: 以場次為單位，顯示「今日/本週」所有訓練的健康度。
*   **關鍵指標**:
    *   **預約進度條**: 顯示 `已預約 / 最大容量`。
    *   **狀態分佈小計**: 快速顯示 `簽到 (Check-in)`、`請假 (Leave)`、`未到 (Pending)` 人數。
*   **交互**: 點擊場次卡片直接進入「簽到詳細頁」。

### 2. 學員數據月報表 (User Monthly Report)
*   **路徑**: `/admin/users/report`
*   **功能**: 以 **家長帳號 (Parent Account)** 為單位彙整營運數據，並支援展開查看 **個別孩子 (Children)** 的明細。
*   **數據分層邏輯**:
    *   **家長層級 (Group Header)**:
        *   彙整該帳號下所有孩子的數據總和（總預約、總出席、總缺席）。
        *   顯示 LINE 帳號名稱與內部 UserID。
    *   **孩子層級 (Expandable Detail)**:
        *   個別孩子的預約、出席、請假、缺席次數。
        *   個別孩子的出席率 (Attendance Rate) 進度條。
*   **數據欄位**:
    *   總預約 (Bookings)、總出席 (Attended)、總請假 (Leave)、總缺席 (Absent)。
    *   出席率 (Attendance Rate): 總出席 / (總預約 - 總請假)。
*   **工具**: 年份/月份篩選、CSV 數據匯出。

### 3. 學員明細與時間軸 (User Drill-down)
*   **路徑**: `/admin/users/:userId`
*   **功能**: 查看特定家長帳號下所有活動的歷史全紀錄。
*   **核心元件**:
    *   **統計摘要**: 該帳號的總體出席表現。
    *   **活動時間軸 (Activity Timeline)**: 
        *   以時間倒序排列紀錄預約、請假、簽到的確切時間點。
        *   包含關聯的孩子姓名與場次資訊。

---

## 三、 專業 UX 設計規範 (Admin UX Guidelines)

### 1. 語意化顏色 (Semantic Coloring)
*   **Blue-400 (`#60A5FA`)**: 確認中、預約中 (Booked)。
*   **Emerald-400 (`#34D399`)**: 正向完成、已簽到 (Checked-in)。
*   **Amber-400 (`#F59E0B`)**: 變動中、已請假 (Leave)。
*   **Red-400 (`#EF4444`)**: 異常、缺席 (Absent)。

### 2. 單手操作優化 (Mobile-Admin First)
*   所有點擊目標（按鈕、Checkbox）高度不小於 **44px**。
*   列表在手機版自動轉為 **卡片佈局 (Card-based layout)**。

---

## 四、 技術實作建議
*   **前端框架**: Go Templ + TailwindCSS + HTMX。
*   **狀態更新**: 透過 HTMX 局部刷新報表與場次卡片，提升管理流暢度。
*   **權限**: 必須通過 `/follow/is-admin` 驗證。
