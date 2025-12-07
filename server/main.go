package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"elotus_test/server/cmd"
	"elotus_test/server/env"
	"elotus_test/server/logger"
	"elotus_test/server/models"
	"elotus_test/server/renv"
)

var cmdFlag = flag.String("cmd", "", "Command mode")
var db = flag.String("db", "", "Database command: migrate, rollback, generate, status")
var migrationName = flag.String("name", "", "Migration name (for generate)")
var steps = flag.Int("steps", 1, "Number of migrations to rollback")

func main() {
	flag.Parse()

	var envConfig *env.ENV
	renv.ParseCmd(&envConfig)
	envConfig.SetDefaults()
	env.E = envConfig

	if env.E.IsDevelopment() {
		logger.InitDevelopment()
	} else {
		logger.InitProduction()
	}

	logger.Info("Starting elotus_test...")
	logger.Infof("Environment: %s", env.E.Environment)
	logger.Infof("Server Name: %s", env.E.ServerName)

	if *db != "" {
		cmd.HandleDB(*db, *migrationName, *steps)
		return
	}

	if *cmdFlag != "" {
		instance := models.NewModels(true)
		instance.RunCmd(*cmdFlag)
		return
	}

	instance := models.NewModels(false)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	logger.Infof("Received signal: %v", sig)
	logger.Info("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := instance.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server shutdown complete")
}
