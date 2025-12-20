package app

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/MaKcm14/price-service/internal/config"
	"github.com/MaKcm14/price-service/internal/controller/chttp"
	"github.com/MaKcm14/price-service/internal/repository/api"
	"github.com/MaKcm14/price-service/internal/repository/api/mmega"
	"github.com/MaKcm14/price-service/internal/repository/api/wildb"
	"github.com/MaKcm14/price-service/internal/repository/kafka"
	"github.com/MaKcm14/price-service/internal/services"
	"github.com/MaKcm14/price-service/internal/services/filter"
	"github.com/MaKcm14/price-service/pkg/entities"
)

// Service unions every parts of the application.
type Service struct {
	appContr    chttp.Controller
	chrome      services.Driver
	producer    services.AsyncWriter
	logger      *slog.Logger
	mainLogFile *os.File
	appSet      config.Settings
}

func NewService() Service {
	log, mainLogFile := setLogger()
	log.Info("main application's configuring begun")

	appSet, err := config.NewSettings(log,
		config.Socket,
		config.ByPassSocket,
		config.Brokers,
		config.AsyncMode,
	)
	if err != nil {
		mainLogFile.Close()
		panic(err)
	}

	var producer kafka.Producer
	if appSet.AsyncMode {
		producer, err = kafka.NewProducer(log, appSet.Brokers)
		if err != nil {
			mainLogFile.Close()
			panic(err)
		}
	}
	chrome := api.NewChromePull()

	return Service{
		appContr: chttp.NewController(
			chttp.ControllerConfig{
				AsyncMode: appSet.AsyncMode,
			},
			echo.New(),
			log,
			filter.New(
				log,
				map[entities.Market]services.ApiInteractor{
					entities.Wildberries: wildb.NewWildberriesAPI(
						chrome.NewContext(), log, 1),
					entities.MegaMarket: mmega.NewMegaMarketAPI(
						chrome.NewContext(), log, appSet.ByPassSocket),
				}, producer)),
		logger:      log,
		mainLogFile: mainLogFile,
		chrome:      chrome,
		appSet:      appSet,
		producer:    producer,
	}
}

// Run starts the configured application.
func (s Service) Run() {
	defer s.mainLogFile.Close()
	defer s.producer.Close()
	defer s.chrome.Close()
	defer s.logger.Info("the app was STOPPED")
	defer s.appContr.GracefulStop()

	s.logger.Info("the app was STARTED")
	go s.appContr.Run(s.appSet.Socket)

	sig := make(chan os.Signal, 3)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig
}

func setLogger() (*slog.Logger, *os.File) {
	date := strings.Split(time.Now().String()[:19], " ")

	mainLogFile, err := os.Create(
		fmt.Sprintf("../../logs/price-service-main-logs_%s___%s.txt",
			date[0],
			strings.Join(
				strings.Split(date[1], ":"),
				"-"),
		))
	if err != nil {
		panic(fmt.Sprintf("error of creating the main-log-file: %v", err))
	}

	log := slog.New(
		slog.NewTextHandler(mainLogFile, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)
	return log, mainLogFile
}
