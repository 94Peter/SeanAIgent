/**
 * 預約狀態校準腳本
 * 目的：將舊版布林值欄位 (is_checked_in, is_on_leave) 的狀態同步到新版字串欄位 (status)
 */

print("🚀 開始執行預約狀態校準...");

// 1. 修正出席狀態：is_checked_in 為 true 但 status 還是 CONFIRMED 的
const attendedResult = db.appointment.updateMany(
    { 
        is_checked_in: true, 
        status: "CONFIRMED" 
    },
    { 
        $set: { 
            status: "ATTENDED", 
            update_at: new Date() 
        } 
    }
);
print("✅ 出席校準完成：已修正 " + attendedResult.modifiedCount + " 筆資料。");

// 2. 修正請假狀態：is_on_leave 為 true 但 status 還是 CONFIRMED 的
const leaveResult = db.appointment.updateMany(
    { 
        is_on_leave: true, 
        status: "CONFIRMED" 
    },
    { 
        $set: { 
            status: "CANCELLED_LEAVE", 
            update_at: new Date() 
        } 
    }
);
print("✅ 請假校準完成：已修正 " + leaveResult.modifiedCount + " 筆資料。");

// 3. 修正缺席狀態：如果 V1 欄位有特定的缺席標記
// (備註：目前缺席主要由 Cron 判定，如果您的歷史資料有特殊缺席標記可在此擴充)

print("
🎉 校準腳本執行完畢。");
print("提示：請記得執行 /cron/sync-all-stats 重新整理報表數據。");
