package main

import (
	"flag"
	"fmt"
	"interview/pkg/db"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"interview/internal/config"
	"interview/internal/router"
	"interview/internal/utils"
	"interview/pkg/log"
)

// Version indicates the current version of the application.
var Version = "1.0.0"

var flagConfig = flag.String("config", "production.yml", "path to the config file")

func main() {
	flag.Parse()

	// create root logger tagged with server version
	logger := log.New().With(nil, "version", Version)

	// load application configurations
	cfg, err := config.Load(*flagConfig, logger)
	if err != nil {
		logger.Errorf("failed to load application configuration: %s", err)
		os.Exit(-1)
	}

	// Open the connection to the database
	dbConnection, err := utils.GetDBConnection(cfg.DSN)
	if err != nil {
		logger.Error(err)
		os.Exit(-1)
	}
	defer func() {
		err := utils.CloseDBConnection(dbConnection)
		if err != nil {
			logger.Error(err)
		}
	}()
	dbctx := db.New(dbConnection, logger)

	// Migrate the database
	err = dbctx.MigrateDatabase()
	if err != nil {
		logger.Error(err)
		os.Exit(-1)
	}

	ginEngine := gin.Default()
	routes := router.New(ginEngine)
	routes.RegisterHandlers(logger, dbctx)

	address := fmt.Sprintf(":%v", cfg.ServerPort)
	srv := &http.Server{
		Addr:    address,
		Handler: ginEngine,
	}

	logger.Infof("server %v is running at %v", Version, address)
	srv.ListenAndServe()
}
