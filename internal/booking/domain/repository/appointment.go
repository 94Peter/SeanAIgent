package repository

import (
	"context"
	"encoding/base64"
	"errors"
	"seanAIgent/internal/booking/domain/entity"
	"strconv"
	"strings"
	"time"
)

type FilterApptsWithTrainDateByCursor struct {
	LastStartDate time.Time
	LastID        string
	PageSize      int
}

func NewFilterApptsWithTrainDateByCursor(pageSize int) string {
	cursor := &FilterApptsWithTrainDateByCursor{
		PageSize: pageSize,
	}
	return cursor.Encode()
}

// 將結構體轉為 Base64 字串
func (c *FilterApptsWithTrainDateByCursor) Encode() string {
	str := strconv.FormatInt(c.LastStartDate.UnixNano(), 10) + "|" + c.LastID + "|" + strconv.Itoa(c.PageSize)
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// 將 Base64 字串轉為結構體
func (c *FilterApptsWithTrainDateByCursor) Decode(cursorStr string) error {
	if cursorStr == "" {
		return errors.New("cursor string is empty")
	}
	// 解碼 Base64 字串
	data, err := base64.StdEncoding.DecodeString(cursorStr)
	if err != nil {
		return err
	}
	parts := strings.Split(string(data), "|")
	if len(parts) != 3 {
		return errors.New("cursor string is invalid")
	}

	// 手動轉換各個欄位
	nano, _ := strconv.ParseInt(parts[0], 10, 64)
	idStr := parts[1]
	pageSize, _ := strconv.ParseInt(parts[2], 10, 64)

	c.LastStartDate = time.Unix(0, nano)
	c.LastID = idStr
	c.PageSize = int(pageSize)
	return nil
}

type AppointmentRepository interface {
	SaveAppointment(ctx context.Context, appt *entity.Appointment) RepoError
	SaveManyAppointments(ctx context.Context, appts []*entity.Appointment) RepoError
	DeleteAppointment(ctx context.Context, appt *entity.Appointment) RepoError
	UpdateAppt(ctx context.Context, appt *entity.Appointment) RepoError
	UpdateManyAppts(ctx context.Context, appts []*entity.Appointment) RepoError

	FindApptByID(ctx context.Context, id string) (*entity.Appointment, RepoError)
	FindApptsByFilter(ctx context.Context, filter FilterAppointment) ([]*entity.Appointment, RepoError)

	PageFindApptsWithTrainDateByFilterAndTrainFilter(
		ctx context.Context,
		filter FilterAppointment,
		trainFilter FilterTrainDate,
		cursorStr string,
	) ([]*entity.AppointmentWithTrainDate, string, RepoError)
}

// query repo filter definition
type FilterAppointment interface {
	isCriteria() // 標記用介面
}

func NewFilterApptByTrainID(id string) FilterAppointment {
	return FilterApptByTrainID{TrainingID: id}
}

type FilterApptByTrainID struct {
	TrainingID string
}

func (f FilterApptByTrainID) isCriteria() {}

func NewFilterApptByIDs(ids ...string) FilterAppointment {
	return FilterApptByIDs{ApptIDs: ids}
}

type FilterApptByIDs struct {
	ApptIDs []string
}

func (f FilterApptByIDs) isCriteria() {}

func NewFilterApptByUserID(userID string) FilterAppointment {
	return FilterAppointmentByUserID{UserID: userID}
}

type FilterAppointmentByUserID struct {
	UserID string
}

func (f FilterAppointmentByUserID) isCriteria() {}
