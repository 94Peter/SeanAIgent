package core

type Kind string

type ErrorType string

const (
	KindDomain  Kind = "DOMAIN"
	KindDB      Kind = "DB"
	KindUseCase Kind = "USE_CASE"

	ErrInternal         ErrorType = "INTERNAL_ERROR"
	ErrInvalidInput     ErrorType = "INVALID_INPUT"
	ErrNotFound         ErrorType = "NOT_FOUND"
	ErrConflict         ErrorType = "BUSINESS_CONFLICT"
	ErrPermissionDenied ErrorType = "PERMISSION_DENIED"
	ErrForbidden        ErrorType = "FORBIDDEN"
)

type UseCaseError interface {
	error
	Unwrap() error
	Kind() Kind
	Type() ErrorType
	Category() string
	Code() string
	Message() string
	Wrap(err error) UseCaseError
}

func newError(kind Kind, category, code, msg string, typ ErrorType) UseCaseError {
	return &useCaseError{
		kind:     kind,
		category: category,
		code:     code,
		message:  msg,
		typ:      typ,
	}
}

type useCaseError struct {
	err      error
	kind     Kind
	category string
	code     string
	message  string
	typ      ErrorType
}

func (e *useCaseError) Error() string {
	return e.message
}

func (e *useCaseError) Unwrap() error {
	return e.err
}

func (e *useCaseError) Wrap(err error) UseCaseError {
	copyErr := *e
	if err != nil {
		copyErr.err = err
	}
	return &copyErr
}

func (e *useCaseError) Kind() Kind {
	return e.kind
}

func (e *useCaseError) Category() string {
	return e.category
}

func (e *useCaseError) Code() string {
	return e.code
}

func (e *useCaseError) Type() ErrorType {
	return e.typ
}

func (e *useCaseError) Message() string {
	return e.message
}

// Domain Error
func NewDomainError(category, code, msg string, typ ErrorType) UseCaseError {
	return newError(KindDomain, category, code, msg, typ)
}

// DB Error
func NewDBError(category, code, msg string, typ ErrorType) UseCaseError {
	return newError(KindDB, category, code, msg, typ)
}

// Server / Infra Error
func NewUseCaseError(category, code, msg string, typ ErrorType) UseCaseError {
	return newError(KindUseCase, category, code, msg, typ)
}
