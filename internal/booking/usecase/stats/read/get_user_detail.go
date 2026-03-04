package read

import (
	"context"
	"fmt"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
	"sort"
)

type ReqGetUserDetail struct {
	UserID string
}

type RespGetUserDetail struct {
	UserID          string
	LineDisplayName string
	OverallStats    UserOverallStatsVO
	MonthlyRecords  []*UserMonthlyRecordVO
}

type UserOverallStatsVO struct {
	TotalBookings  int
	TotalAttended  int
	TotalLeave     int
	TotalAbsent    int
	AttendanceRate float64
}

type UserMonthlyRecordVO struct {
	MonthDisplay string
	MonthValue   string
	Bookings     []*ChildBookingRecordVO
}

type ChildBookingRecordVO struct {
	ChildName string
	Date      string
	Time      string
	Location  string
	Status    string
}

type GetUserDetailUseCase core.ReadUseCase[ReqGetUserDetail, *RespGetUserDetail]

type getUserDetailUseCase struct {
	repo repository.StatsRepository
}

func NewGetUserDetailUseCase(repo repository.StatsRepository) GetUserDetailUseCase {
	return &getUserDetailUseCase{repo: repo}
}

func (uc *getUserDetailUseCase) Name() string {
	return "GetUserDetail"
}

func (uc *getUserDetailUseCase) Execute(ctx context.Context, req ReqGetUserDetail) (*RespGetUserDetail, core.UseCaseError) {
	// 1. 獲取用戶所有預約明細 (這是最準確的原始數據來源)
	allTimeAppts, err := uc.repo.GetUserApptStats(ctx, req.UserID, nil)
	if err != nil {
		return nil, core.NewDBError("GET_USER_DETAIL", "FETCH_APPTS_FAIL", "failed to fetch appointments", core.ErrInternal).Wrap(err)
	}

	resp := &RespGetUserDetail{
		UserID:          req.UserID,
		LineDisplayName: allTimeAppts.UserName,
		MonthlyRecords:  make([]*UserMonthlyRecordVO, 0),
	}

	// 2. 計算全時段累績統計
	// 直接從聚合結果中的全體計數拿
	resp.OverallStats = UserOverallStatsVO{
		TotalBookings: allTimeAppts.TotalAppointment,
		TotalAttended: allTimeAppts.CheckedInCount,
		TotalLeave:    allTimeAppts.OnLeaveCount,
	}
	
	// 重新計算缺席數 (由於 UserApptStats 沒存總缺席，我們手動從明細算)
	var totalAbs int
	monthlyMap := make(map[string][]*ChildBookingRecordVO)
	
	for _, child := range allTimeAppts.ChildState {
		for _, appt := range child.Appointments {
			monthKey := appt.StartDate.Format("2006-01")
			status := "Pending"
			if appt.IsCheckedIn {
				status = "CheckedIn"
			} else if appt.IsOnLeave {
				status = "Leave"
			} else if appt.IsAbsent {
				status = "Absent"
				totalAbs++
			}
			
			monthlyMap[monthKey] = append(monthlyMap[monthKey], &ChildBookingRecordVO{
				ChildName: child.ChildName,
				Date:      appt.StartDate.Format("01/02 (週一)"), 
				Time:      appt.StartDate.Format("15:04"),
				Location:  appt.Location,
				Status:    status,
			})
		}
	}
	resp.OverallStats.TotalAbsent = totalAbs
	if resp.OverallStats.TotalBookings > 0 {
		resp.OverallStats.AttendanceRate = float64(resp.OverallStats.TotalAttended) / float64(resp.OverallStats.TotalBookings)
	}

	// 3. 組織每月明細列表並排序 (由新到舊)
	var monthKeys []string
	for k := range monthlyMap {
		monthKeys = append(monthKeys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(monthKeys)))

	for _, k := range monthKeys {
		var y, m int
		fmt.Sscanf(k, "%d-%d", &y, &m)
		resp.MonthlyRecords = append(resp.MonthlyRecords, &UserMonthlyRecordVO{
			MonthDisplay: fmt.Sprintf("%d年 %d月", y, m),
			MonthValue:   k,
			Bookings:     monthlyMap[k],
		})
	}

	return resp, nil
}
