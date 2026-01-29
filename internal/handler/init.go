package handler

import (
	"seanAIgent/internal/service"

	"github.com/94peter/vulpes/ezapi"
)

func InitHandler(routers ezapi.RouterGroup, service service.Service, enableCSRF bool) {
	// v1Router := routers.Group("v1")
	initBookingApi(routers, service, enableCSRF)
	initHealthApi(routers)
	initCheckinApi(routers, service, enableCSRF)
	initTrainingApi(routers, service, enableCSRF)
	initComponentApi(routers)
}
