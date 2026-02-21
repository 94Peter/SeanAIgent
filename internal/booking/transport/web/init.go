package web

import (
	"context"
	"net/http"
	"seanAIgent/internal/booking/transport/web/handler"
	"seanAIgent/locales"

	"github.com/94peter/botreplyer"
	lineMid "github.com/94peter/botreplyer/provider/line/mid"
	"github.com/94peter/vulpes/ezapi"
	"github.com/94peter/vulpes/log"
)

type WebService interface {
	Run(ctx context.Context)
}

func getApis(
	enableCSRF bool,
	bookingUseCaseSet handler.BookingUseCaseSet,
	trainingUseCaseSet handler.TrainingUseCaseSet,
	v2BookingUseCaseSet handler.V2BookingUseCaseSet,
) []handler.WebAPI {
	return []handler.WebAPI{
		handler.NewBookingApi(enableCSRF, bookingUseCaseSet),
		handler.NewTrainingApi(enableCSRF, trainingUseCaseSet),
		handler.NewHealthApi(),
		handler.NewComponentApi(),
		handler.NewV2BookingApi(enableCSRF, v2BookingUseCaseSet),
	}
}

func InitWeb(
	bookingUseCaseSet handler.BookingUseCaseSet,
	trainingUseCaseSet handler.TrainingUseCaseSet,
	v2BookingUseCaseSet handler.V2BookingUseCaseSet,
	cfg Config,
) WebService {
	router := ezapi.NewRouterGroup()
	// v2Router := router.Group("v2")
	apis := getApis(
		cfg.csrf.Enable,
		bookingUseCaseSet,
		trainingUseCaseSet,
		v2BookingUseCaseSet,
	)
	for _, api := range apis {
		api.InitRouter(router)
	}
	cfg.routerGroup = router
	return &webService{
		cfg: cfg,
	}
}

type Config struct {
	routerGroup ezapi.RouterGroup
	csrf        struct {
		Secret       string
		FieldName    string
		ExcludePaths []string
		Enable       bool
	}
	session struct {
		Store      string
		CookieName string
		KeyPairs   []string
		MaxAge     int
		Enable     bool
	}
	logger struct {
		Enable bool
	}
	tracer struct {
		Enable bool
	}
	mode                  string
	port                  uint16
	maxConcurrentRequests int
}

type webService struct {
	cfg Config
}

func (s *webService) Run(ctx context.Context) {
	if err := ezapi.RunGin(
		ctx,
		ezapi.WithRouterGroup(s.cfg.routerGroup),
		ezapi.WithPort(s.cfg.port),
		ezapi.WithMiddleware(
			lineMid.LineLiff(),
			ezapi.I18n("zh-tw", locales.LocaleExist),
			lineMid.CheckAdmin(botreplyer.GetFollowStore()),
		),
		ezapi.WithSession(
			s.cfg.session.Enable,
			s.cfg.session.Store,
			s.cfg.session.CookieName,
			s.cfg.session.MaxAge,
			s.cfg.session.KeyPairs...,
		),
		ezapi.WithCsrf(
			s.cfg.csrf.Enable,
			s.cfg.csrf.Secret,
			s.cfg.csrf.FieldName,
			s.cfg.csrf.ExcludePaths...,
		),
		ezapi.WithStaticFS("/assets", http.Dir("assets")),
		ezapi.WithTracerEnable(s.cfg.tracer.Enable),
		ezapi.WithLoggerEnable(s.cfg.logger.Enable),
		ezapi.WithMode(s.cfg.mode),
		ezapi.WithMaxConcurrentRequests(s.cfg.maxConcurrentRequests),
	); err != nil {
		log.Info(err.Error())
	}
}
