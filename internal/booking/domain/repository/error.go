package repository

import (
	"errors"
	"strings"
)

var (
	// ErrNotFound 當查詢不到預期的資料時回傳
	ErrNotFound = errors.New("repository: resource not found")

	// ErrConflict 當資料違反唯一約束時（如重複預約）回傳
	ErrConflict = errors.New("repository: data conflict")

	// ErrInternal 當資料庫發生非預期的連線失敗或系統錯誤時回傳
	ErrInternal             = errors.New("repository: internal database error")
	ErrFilterNotImplemented = errors.New("repository: filter not implemented")
	ErrInvalidDocumentID    = errors.New("repository: invalid document id")
	ErrInvalidCursor        = errors.New("repository: invalid cursor")
)

func NewRepoInvalidDocumentIDError(repo, dbType, op string, rawErr error) RepoError {
	return &repoError{
		repo:      repo,
		dbType:    dbType,
		op:        op,
		rawErr:    rawErr,
		domainErr: ErrInvalidDocumentID,
	}
}

func NewRepoNotFoundError(repo, dbType, op string, rawErr error) RepoError {
	return &repoError{
		repo:      repo,
		dbType:    dbType,
		op:        op,
		rawErr:    rawErr,
		domainErr: ErrNotFound,
	}
}

func NewRepoConflictError(repo, dbType, op string, rawErr error) RepoError {
	return &repoError{
		repo:      repo,
		dbType:    dbType,
		op:        op,
		rawErr:    rawErr,
		domainErr: ErrConflict,
	}
}

func NewRepoInternalError(repo, dbType, op string, rawErr error) RepoError {
	return &repoError{
		repo:      repo,
		dbType:    dbType,
		op:        op,
		rawErr:    rawErr,
		domainErr: ErrInternal,
	}
}

func NewRepoInvalidCursorError(repo, dbType, op string, rawErr error) RepoError {
	return &repoError{
		repo:      repo,
		dbType:    dbType,
		op:        op,
		rawErr:    rawErr,
		domainErr: ErrInvalidCursor,
	}
}

type RepoError interface {
	error
	Unwrap() error
}

type repoError struct {
	rawErr    error
	domainErr error
	repo      string
	dbType    string
	op        string
}

func (e *repoError) Error() string {
	var b strings.Builder
	// 預估長度：[DB/Repo] Op failed: + ErrString
	// 可以根據你的 Repo 名稱長度大約預估一個數值
	b.Grow(64 + len(e.rawErr.Error()))

	b.WriteByte('[')
	b.WriteString(e.dbType)
	b.WriteByte('/')
	b.WriteString(e.repo)
	b.WriteString("] ")
	b.WriteString(e.op)
	b.WriteString(" failed: ")
	if e.rawErr != nil {
		b.WriteString(e.rawErr.Error())
	}

	return b.String()
}

// 關鍵在於 Unwrap
func (e *repoError) Unwrap() error {
	// 這裡回傳 DomainErr，讓 errors.Is(err, repository.ErrNotFound) 成立
	return e.domainErr
}
