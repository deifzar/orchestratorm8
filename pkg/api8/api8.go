package api8

import (
	"database/sql"
	"deifzar/orchestratorm8/pkg/cleanup8"
	"deifzar/orchestratorm8/pkg/configparser"
	"deifzar/orchestratorm8/pkg/controller8"
	"deifzar/orchestratorm8/pkg/db8"
	"deifzar/orchestratorm8/pkg/log8"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/spf13/viper"
)

type Api8 struct {
	Cnfg   *viper.Viper
	DB     *sql.DB
	Router *gin.Engine
}

func (a *Api8) Init() error {

	// Create configs, log and tmp directories if they don't exist
	if err := os.MkdirAll("configs", 0755); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to create configs directory")
		return err
	}
	if err := os.MkdirAll("log", 0755); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to create log directory")
		return err
	}
	if err := os.MkdirAll("tmp", 0755); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to create tmp directory")
		return err
	}

	// Clean up old files in tmp directory (older than 24 hours)
	cleanup := cleanup8.NewCleanup8()
	if err := cleanup.CleanupDirectory("tmp", 24*time.Hour); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to cleanup tmp directory")
		// Don't return error here as cleanup failure shouldn't prevent startup
	}

	v, err := configparser.InitConfigParser()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("Error initialising the config parser.")
		return err
	}
	location := v.GetString("Database.location")
	port := v.GetInt("Database.port")
	schema := v.GetString("Database.schema")
	database := v.GetString("Database.database")
	username := v.GetString("Database.username")
	password := v.GetString("Database.password")

	var db db8.Db8
	db.InitDatabase8(location, port, schema, database, username, password)
	conn, err := db.OpenConnection()
	if err != err {
		log8.BaseLogger.Error().Msg("Error connecting into DB.")
		return err
	}

	a.Cnfg = v
	a.DB = conn
	return nil
}

func (a *Api8) Routes() (controller8.Controller8OrchestratorM8Interface, error) {
	r := gin.Default()
	contrOrch, err := controller8.NewController8OrchestratorM8(a.DB, a.Cnfg)
	if err != nil {
		log8.BaseLogger.Info().Msg("Error initializing the default RabbitMQ pool.")
		return &controller8.Controller8OrchestratorM8{}, err
	}

	err = contrOrch.InitOrchestrator()
	if err != nil {
		log8.BaseLogger.Info().Msg("Error declaring exchange and queues in RabbitMQ.")
		return &controller8.Controller8OrchestratorM8{}, err
	}

	r.MaxMultipartMemory = 8 << 20 //8 MiB

	// Health live probes
	r.GET("/health", contrOrch.HealthCheck)
	r.GET("/ready", contrOrch.ReadinessCheck)

	a.Router = r
	return contrOrch, nil
}

func (a *Api8) Run(addr string) {
	a.Router.Run(addr)
}
