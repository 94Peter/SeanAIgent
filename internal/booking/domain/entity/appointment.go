package entity

import (
	"errors"
	"fmt"
	"time"
)

type appointmentStatus string
type leaveStatus string

const (
	StatusConfirmed appointmentStatus = "CONFIRMED" // 已預約
	StatusAttended  appointmentStatus = "ATTENDED"  // 已出席
	// StatusAbsent         appointmentStatus = "ABSENT"          // 缺席
	StatusCancelled      appointmentStatus = "CANCELLED"       // 10分鐘內誤按取消
	StatusCancelledLeave appointmentStatus = "CANCELLED_LEAVE" // 已請假

	LeaveStatusNone     leaveStatus = "none" // No leave request or not applicable
	LeaveStatusPending  leaveStatus = "pending"
	LeaveStatusApproved leaveStatus = "approved"
	LeaveStatusRejected leaveStatus = "rejected"
)

var (
	apptStatusTrans = map[string]appointmentStatus{
		string(StatusConfirmed): StatusConfirmed,
		string(StatusAttended):  StatusAttended,
		// string(StatusAbsent):         StatusAbsent,
		// string(StatusCancelled):      StatusCancelled,
		string(StatusCancelledLeave): StatusCancelledLeave,
	}

	leaveStatusTrans = map[string]leaveStatus{
		string(LeaveStatusNone):     LeaveStatusNone,
		string(LeaveStatusPending):  LeaveStatusPending,
		string(LeaveStatusApproved): LeaveStatusApproved,
		string(LeaveStatusRejected): LeaveStatusRejected,
	}

	EmptyLeaveInfo = VO_LeaveInfo{
		reason:    "",
		status:    LeaveStatusNone,
		createdAt: time.Time{},
	}
)

func LeaveStatusFromString(status string) (leaveStatus, bool) {
	leaveStatus, ok := leaveStatusTrans[status]
	return leaveStatus, ok
}

func AppointmentStatusFromString(status string) (appointmentStatus, bool) {
	apptStatus, ok := apptStatusTrans[status]
	return apptStatus, ok
}

type Appointment struct {
	leave      VO_LeaveInfo
	createdAt  time.Time
	updateAt   time.Time
	verifiedAt *time.Time
	user       User
	id         string
	childName  string
	trainingId string
	status     appointmentStatus
}

func NewLeaveInfo(reason string, status leaveStatus, createdAt time.Time) VO_LeaveInfo {
	return VO_LeaveInfo{
		reason:    reason,
		status:    status,
		createdAt: createdAt,
	}
}

type VO_LeaveInfo struct {
	createdAt time.Time
	reason    string
	status    leaveStatus
}

func (l VO_LeaveInfo) IsEmpty() bool {
	return l.status == LeaveStatusNone
}

func (a Appointment) Validate() error {
	if a.childName == "" {
		return fmt.Errorf("%w: child name is empty", ErrAppointmentInvalid)
	}
	if a.trainingId == "" {
		return fmt.Errorf("%w: training id is empty", ErrAppointmentInvalid)
	}
	if a.id == "" {
		return fmt.Errorf("%w: id is empty", ErrAppointmentInvalid)
	}
	if a.status == "" {
		return fmt.Errorf("%w: status is empty", ErrAppointmentInvalid)
	}
	if a.createdAt.IsZero() {
		return fmt.Errorf("%w: created at is zero", ErrAppointmentInvalid)
	}
	if a.updateAt.IsZero() {
		return fmt.Errorf("%w: update at is zero", ErrAppointmentInvalid)
	}
	return nil
}

type apptOpt func(appt *Appointment)

func WithCreateAppt(id, trainingID string, user User, childName string) apptOpt {
	return func(appt *Appointment) {
		appt.id = id
		appt.trainingId = trainingID
		appt.user = user
		appt.childName = childName
		appt.status = StatusConfirmed
		appt.createdAt = time.Now()
		appt.updateAt = appt.createdAt
	}
}

func WithApptID(id string) apptOpt {
	return func(appt *Appointment) {
		appt.id = id
	}
}

func WithUser(u User) apptOpt {
	return func(appt *Appointment) {
		appt.user = u
	}
}

func WithChildName(name string) apptOpt {
	return func(appt *Appointment) {
		appt.childName = name
	}
}

func WithTrainingID(id string) apptOpt {
	return func(appt *Appointment) {
		appt.trainingId = id
	}
}

func WithStatus(status appointmentStatus) apptOpt {
	return func(appt *Appointment) {
		appt.status = status
	}
}

func WithCreatedAt(t time.Time) apptOpt {
	return func(appt *Appointment) {
		appt.createdAt = t
	}
}

func WithUpdatedAt(t time.Time) apptOpt {
	return func(appt *Appointment) {
		appt.updateAt = t
	}
}

func WithVerifiedAt(t *time.Time) apptOpt {
	return func(appt *Appointment) {
		appt.verifiedAt = t
	}
}

func WithLeaveInfo(leave VO_LeaveInfo) apptOpt {
	return func(appt *Appointment) {
		appt.leave = leave
	}
}

func NewAppointment(opts ...apptOpt) (*Appointment, error) {
	newAppt := &Appointment{
		leave: EmptyLeaveInfo,
	}
	for _, opt := range opts {
		opt(newAppt)
	}
	if err := newAppt.Validate(); err != nil {
		return nil, err
	}
	return newAppt, nil
}

const cancelTimeout = 24 * time.Hour

func (a *Appointment) CancelAsMistake(userID string) error {
	if a.user.userID != userID {
		return ErrAppointmentNotBelongToUser
	}
	if time.Since(a.createdAt) > cancelTimeout {
		return ErrAppointmentCancelTimeout
	}
	if a.status != StatusConfirmed {
		return ErrAppointmentInvalidStatus
	}

	a.status = StatusCancelled
	a.updateAt = time.Now()
	return nil
}

func (a *Appointment) MarkAsAttended(trainingStartTime time.Time) error {
	// 規則：上課前 10 分鐘才開放點名
	if time.Now().Before(trainingStartTime.Add(-10 * time.Minute)) {
		return ErrAppointmentCheckInNotOpen
	}
	// 規則：課程結束3天內可以再補簽到
	if time.Now().After(trainingStartTime.Add(3 * 24 * time.Hour)) {
		return ErrAppointmentCheckInTooLate
	}
	// 規則：請假中無法簽到
	if a.status == StatusCancelledLeave {
		return ErrAppointmentOnLeave
	}

	now := time.Now()
	a.status = StatusAttended
	a.verifiedAt = &now
	a.updateAt = now
	return nil
}

func (a *Appointment) AppendLeaveRecord(reason string, trainingStartTime time.Time) error {
	if reason == "" {
		return ErrAppointmentLeaveReasonEmpty
	}
	// 1. 檢查規則：必須在 2 小時前
	if time.Until(trainingStartTime) < 2*time.Hour {
		return ErrAppointmentLeaveTooLate
	}

	if a.status != StatusConfirmed {
		return ErrAppointmentCannotLeave
	}

	// 2. 更新自身狀態
	a.status = StatusCancelledLeave

	a.leave = VO_LeaveInfo{
		reason:    reason,
		status:    LeaveStatusApproved,
		createdAt: time.Now(),
	}
	a.updateAt = time.Now()

	return nil
}

func (a *Appointment) CancelLeave(userID string) error {
	if a.user.userID != userID {
		return ErrAppointmentNotBelongToUser
	}
	if a.leave.status != LeaveStatusApproved && a.leave.status != LeaveStatusPending {
		return ErrAppointmentLeaveNotApproved
	}
	a.leave = EmptyLeaveInfo
	a.status = StatusConfirmed
	a.updateAt = time.Now()
	return nil
}

// Error Definition
var (
	ErrAppointmentCheckInTooLate   = errors.New("APPOINTMENT_CHECKIN_TOO_LATE")
	ErrAppointmentNotBelongToUser  = errors.New("APPOINTMENT_NOT_BELONG_TO_USER")
	ErrAppointmentInvalid          = errors.New("APPOINTMENT_INVALID")
	ErrAppointmentCancelTimeout    = errors.New("APPOINTMENT_CANCEL_TIMEOUT")
	ErrAppointmentInvalidStatus    = errors.New("APPOINTMENT_INVALID_STATUS")
	ErrAppointmentCheckInNotOpen   = errors.New("APPOINTMENT_CHECKIN_NOT_OPEN")
	ErrAppointmentOnLeave          = errors.New("APPOINTMENT_ON_LEAVE")
	ErrAppointmentLeaveTooLate     = errors.New("APPOINTMENT_LEAVE_TOO_LATE")
	ErrAppointmentCannotLeave      = errors.New("APPOINTMENT_CANNOT_LEAVE")
	ErrAppointmentLeaveReasonEmpty = errors.New("APPOINTMENT_LEAVE_REASON_EMPTY")
	ErrAppointmentLeaveNotApproved = errors.New("APPOINTMENT_LEAVE_NOT_APPROVED")
)

// Getter
func (appt *Appointment) ID() string {
	return appt.id
}

func (appt *Appointment) User() User {
	return appt.user
}

func (appt *Appointment) ChildName() string {
	return appt.childName
}

func (appt *Appointment) TrainingID() string {
	return appt.trainingId
}
func (appt *Appointment) Status() appointmentStatus {
	return appt.status
}
func (appt *Appointment) CreatedAt() time.Time {
	return appt.createdAt
}
func (appt *Appointment) UpdateAt() time.Time {
	return appt.updateAt
}
func (appt *Appointment) VerifiedAt() *time.Time {
	return appt.verifiedAt
}

func (appt *Appointment) LeaveInfo() VO_LeaveInfo {
	return appt.leave
}

func (leave VO_LeaveInfo) Reason() string {
	return leave.reason
}

func (leave VO_LeaveInfo) Status() leaveStatus {
	return leave.status
}

func (leave VO_LeaveInfo) CreatedAt() time.Time {
	return leave.createdAt
}
