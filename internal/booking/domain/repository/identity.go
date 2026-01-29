package repository

type IdentityGenerator interface {
	GenerateID() string
}
