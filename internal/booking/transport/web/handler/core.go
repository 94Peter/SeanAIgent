package handler

import "github.com/94peter/vulpes/ezapi"

type WebAPI interface {
	InitRouter(r ezapi.Router)
}
