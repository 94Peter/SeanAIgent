package core

import "seanAIgent/internal/booking/domain/repository"

const dbType = "mongodb"

func NewInvalidDocumentIDError(repo, op string, rawErr error) repository.RepoError {
	return repository.NewRepoInvalidDocumentIDError(repo, dbType, op, rawErr)
}

func NewNotFoundError(repo, op string, rawErr error) repository.RepoError {
	return repository.NewRepoNotFoundError(repo, dbType, op, rawErr)
}

func NewConflictError(repo, op string, rawErr error) repository.RepoError {
	return repository.NewRepoConflictError(repo, dbType, op, rawErr)
}

func NewInternalError(repo, op string, rawErr error) repository.RepoError {
	return repository.NewRepoInternalError(repo, dbType, op, rawErr)
}

func NewInvalidCursorError(repo, op string, rawErr error) repository.RepoError {
	return repository.NewRepoInvalidCursorError(repo, dbType, op, rawErr)
}
