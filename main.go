package main

import (
	"fmt"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"

	"github.com/gabihodoroaga/cloudrun-casbin/auth"
	"github.com/gabihodoroaga/cloudrun-casbin/config"
	"github.com/gabihodoroaga/cloudrun-casbin/db"
	"github.com/gabihodoroaga/cloudrun-casbin/routes"
)

func main() {

	err := config.LoadConfig()
	if err != nil {
		fmt.Printf("error initializing configuration %v", err)
		os.Exit(1)
	}

	logger, err := initLogger()
	if err != nil {
		fmt.Printf("error initializing logger %v", err)
		os.Exit(1)
	}
	logger.Info("main: start")

	if err := db.SetupDB(); err != nil {
		fmt.Printf("error initializing database %v", err)
		os.Exit(1)
	}

	if err := auth.SetupAuth(); err != nil {
		fmt.Printf("error initializing authorization %v", err)
		os.Exit(1)
	}

	r := gin.New()
	r.Use(logWithZap(logger))
	r.Use(recoveryWithZap(logger, true))

	r.RedirectTrailingSlash = false

	// setup cors
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AddAllowHeaders("Authorization")
	r.Use(cors.New(corsConfig))

	r.Use(static.Serve("/", static.LocalFile("./public", true)))

	routes.SetupRoutes(r)

	r.Run()
}
