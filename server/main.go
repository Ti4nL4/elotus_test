package main

import (
	"flag"

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

	// Parse environment configuration
	var envConfig *env.ENV
	renv.ParseCmd(&envConfig)
	envConfig.SetDefaults()
	env.E = envConfig

	// Initialize logger based on environment
	if env.E.IsDevelopment() {
		logger.InitDevelopment()
	} else {
		logger.InitProduction()
	}

	logger.Info("Starting elotus_test...")
	logger.Infof("Environment: %s", env.E.Environment)
	logger.Infof("Server Name: %s", env.E.ServerName)

	// Handle database commands
	if *db != "" {
		cmd.HandleDB(*db, *migrationName, *steps)
		return
	}

	// Handle other commands
	if *cmdFlag != "" {
		instance := models.NewModels(true)
		instance.RunCmd(*cmdFlag)
		return
	}

	// Start server
	models.NewModels(false)
	select {}
}
