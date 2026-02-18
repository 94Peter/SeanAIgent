# Booking V2 API Specification

此文件定義了 Booking V2 前端頁面所需的後端 API 接口。所有回應皆使用 JSON 格式。

Base Path: `/api/v2`

---

## 1. 建立預約 (Create Booking)

當使用者在 Detail Popup 的「新增參與者」區域輸入姓名並點擊「完成」時呼叫。

- **Method:** `POST`
- **Path:** `/bookings`
- **Request Headers:**
  - `Content-Type: application/json`
  - `X-CSRF-Token`: (如果有啟用 CSRF)

### Request Body
```json
{
  "slot_id": "2026-02-13-slot-1",
  "student_names": ["小明", "小華"]
}
```

### Response (200 OK)
```json
{
  "success": true,
  "message": "預約成功",
  "new_bookings": [
    {
      "booking_id": "b_123",
      "name": "小明",
      "status": "Booked",
      "booking_time": "2026-02-14T10:00:00Z"
    },
    {
      "booking_id": "b_124",
      "name": "小華",
      "status": "Booked",
      "booking_time": "2026-02-14T10:00:00Z"
    }
  ]
}
```

### Error Response (400/500)
```json
{
  "success": false,
  "message": "名額不足或時段已結束"
}
```

---

## 2. 取消預約 (Cancel Booking)

當使用者點擊藍色「已預約」標籤，且在 24 小時寬限期內，選擇「直接取消」時呼叫。

- **Method:** `DELETE`
- **Path:** `/bookings/:booking_id`

### Response (200 OK)
```json
{
  "success": true,
  "message": "預約已取消"
}
```

---

## 3. 提交請假 (Submit Leave Request)

當預約超過 24 小時，使用者填寫請假原因並送出表單時呼叫。

- **Method:** `POST`
- **Path:** `/bookings/:booking_id/leave`

### Request Body
```json
{
  "reason": "身體不適"
}
```

### Response (200 OK)
```json
{
  "success": true,
  "message": "請假申請已送出"
}
```

---

## 4. 取消請假 / 恢復預約 (Cancel Leave)

當使用者點擊紅色「請假」標籤，並確認要恢復預約時呼叫。

- **Method:** `DELETE`
- **Path:** `/bookings/:booking_id/leave`

### Response (200 OK)
```json
{
  "success": true,
  "message": "已取消請假並恢復預約",
  "current_status": "Booked"
}
```

---

## 5. 取得我的預約列表 (Get My Bookings)

用於「我的預約」Modal 的資料載入，支援分頁 (Infinite Scroll)。

- **Method:** `GET`
- **Path:** `/my-bookings`
- **Query Parameters:**
  - `type`: (Optional) `upcoming` (預設, 即將到來) 或 `history` (歷史紀錄)。
  - `cursor`: (Optional) 上一頁的最後一筆資料指標，用於載入下一頁。
  - `limit`: (Optional) 每頁筆數，預設 10。

### Response (200 OK)
```json
{
  "items": [
    {
      "booking_id": "b_123",
      "date_display": "02/14 (六) 14:30",
      "title": "足球進階班 @ 北安球場",
      "attendees": [
         { 
           "name": "小明", 
           "status": "Booked", 
           "booking_time": "2026-02-10T10:00:00Z",
           "booking_id": "b_123"
         }
      ]
    }
  ],
  "next_cursor": "eyJid... (base64 string)",
  "has_more": true
}

---

## 6. 取得行事曆週次 (Get Calendar Weeks)

用於載入更多週次 (上一週/下一週) 的資料。

- **Method:** `GET`
- **Path:** `/calendar/weeks`
- **Query Parameters:**
  - `start_date`: (Required) 基準日期 (YYYY-MM-DD)，通常是目前畫面上第一週或最後一週的日期。
  - `direction`: (Optional) `next` (預設, 往後載入) 或 `prev` (往前載入)。
  - `limit`: (Optional) 載入週數，預設 2。

### Response (200 OK)
```json
{
  "weeks": [
    {
      "id": "2026-W09",
      "days": [
        {
          "date_display": "22",
          "day_of_week": "Sun",
          "is_today": false,
          "full_date": "2026-02-22",
          "slots": [
             { "is_empty": true }
          ]
        }
        // ... more days
      ]
    }
  ]
}
```

---

## 6. 取得行事曆週次 (Get Calendar Weeks)

用於載入更多週次 (上一週/下一週) 的資料。

- **Method:** `GET`
- **Path:** `/calendar/weeks`
- **Query Parameters:**
  - `start_date`: (Required) 基準日期 (YYYY-MM-DD)，通常是目前畫面上第一週或最後一週的日期。
  - `direction`: (Optional) `next` (預設, 往後載入) 或 `prev` (往前載入)。
  - `limit`: (Optional) 載入週數，預設 2。

### Response (200 OK)
```json
{
  "weeks": [
    {
      "id": "2026-W09",
      "days": [
        {
          "date_display": "22",
          "day_of_week": "Sun",
          "is_today": false,
          "full_date": "2026-02-22",
          "slots": [
             { "is_empty": true }
          ]
        }
        // ... more days
      ]
    }
  ]
}
```
```
