package main

import (
	"flag"
	"interview/pkg/controllers"
	"interview/pkg/db"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"interview/internal/config"
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
	dbConnection, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		logger.Error(err)
		os.Exit(-1)
	}
	defer func() {
		dbInstance, _ := dbConnection.DB()
		if err := dbInstance.Close(); err != nil {
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

	var taxController controllers.TaxController
	ginEngine.GET("/", taxController.ShowAddItemForm)
	ginEngine.POST("/add-item", taxController.AddItem)
	ginEngine.GET("/remove-cart-item", taxController.DeleteCartItem)
	srv := &http.Server{
		Addr:    ":8088",
		Handler: ginEngine,
	}

	srv.ListenAndServe()
}
