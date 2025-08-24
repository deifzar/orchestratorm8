package orchestratorm8

import (
	"database/sql"
	amqpM8 "deifzar/orchestratorm8/pkg/amqpM8"
	"deifzar/orchestratorm8/pkg/cleanup8"
	"deifzar/orchestratorm8/pkg/configparser"
	"deifzar/orchestratorm8/pkg/db8"
	"deifzar/orchestratorm8/pkg/log8"
	"strconv"
	"time"

	"github.com/spf13/viper"

	_ "github.com/lib/pq"
)

type Orchestrator8 struct {
	Config *viper.Viper
	Db     *sql.DB
}

func NewOrchestratorM8() (Orchestrator8Interface, error) {

	v, err := configparser.InitConfigParser()
	if err != nil {
		return &Orchestrator8{}, err
	}

	// Try to initialize the default pool (will do nothing if already exists)
	manager := amqpM8.GetGlobalPoolManager()
	poolExists := false
	for _, poolName := range manager.ListPools() {
		if poolName == "default" {
			poolExists = true
			break
		}
	}

	if !poolExists {
		err = amqpM8.InitializeConnectionPool()
		if err != nil {
			log8.BaseLogger.Debug().Msg(err.Error())
			return &Orchestrator8{}, err
		}
	}

	locationDB := v.GetString("Database.location")
	portDB := v.GetInt("Database.port")
	schemaDB := v.GetString("Database.schema")
	databaseDB := v.GetString("Database.database")
	usernameDB := v.GetString("Database.username")
	passwordDB := v.GetString("Database.password")

	var db db8.Db8
	db.InitDatabase8(locationDB, portDB, schemaDB, databaseDB, usernameDB, passwordDB)
	conn, err := db.OpenConnection()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msg("Error connecting into DB.")
		return &Orchestrator8{}, err
	}

	o := &Orchestrator8{Config: v, Db: conn}
	return o, nil
}

func (o *Orchestrator8) InitOrchestrator() error {
	exchanges := o.Config.GetStringMapString("ORCHESTRATORM8.Exchanges")

	queue := o.Config.GetStringSlice("ORCHESTRATORM8.asmm8.Queue")
	bindingkeys := o.Config.GetStringSlice("ORCHESTRATORM8.asmm8.Routing-keys")
	qargs := o.Config.GetStringMap("ORCHESTRATORM8.asmm8.Queue-arguments")
	prefetch_count, err := strconv.Atoi(queue[2])
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return err
	}

	err = amqpM8.WithPooledConnection(func(am8 amqpM8.PooledAmqpInterface) error {
		for exname, extype := range exchanges {
			err := am8.DeclareExchange(exname, extype)
			if err != nil {
				log8.BaseLogger.Debug().Msg(err.Error())
				return err
			}
		}
		// Declare 'ASMM8' queue and bind it to the 'CPTM8' Exchange
		log8.BaseLogger.Info().Msg("RabbitMQ declaring queues for the ASMM8 service.")
		err = am8.DeclareQueue(queue[0], queue[1], prefetch_count, qargs)
		if err != nil {
			log8.BaseLogger.Debug().Msg(err.Error())
			return err
		}
		err = am8.BindQueue(queue[0], queue[1], bindingkeys)
		if err != nil {
			log8.BaseLogger.Debug().Msg(err.Error())
			return err
		}
		return nil
	})

	return err
}

func (o *Orchestrator8) StartOrchestrator() {

	// Clean up old files in tmp directory (older than 24 hours)
	cleanup := cleanup8.NewCleanup8()
	if err := cleanup.CleanupDirectory("tmp", 24*time.Hour); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to cleanup tmp directory")
		// Don't return error here as cleanup failure shouldn't prevent startup
	}

	DB := o.Db
	domain8 := db8.NewDb8Domain8(DB)

	var emptyChan = make(chan bool)
	var publishChan = make(chan bool)

	checkDBEmpty := func(first bool) {
		for {
			domains, err := domain8.GetAllEnabled()
			if err != nil {
				log8.BaseLogger.Debug().Msg(err.Error())
				log8.BaseLogger.Error().Msgf("Starting orchestrator has failed. Something wrong fetching domains from the DB.")
			}
			emptyChan <- (len(domains) < 1)
			if first && (len(domains) > 0) {
				publishChan <- true
				first = false
			} else {
				if len(domains) < 1 {
					publishChan <- false
					first = true
				} else {
					publishChan <- false
					first = false
				}
			}
			time.Sleep(15 * time.Minute)
		}
	}
	checkRMQPublish := func() {
		for {
			empty := <-emptyChan
			publish := <-publishChan
			// if there is at least one domain and you can publish
			if !empty {
				if publish {
					log8.BaseLogger.Info().Msg("There are domains in the DB and publishing messages is allowed.")
					queue := o.Config.GetStringSlice("ORCHESTRATORM8.asmm8.Queue")
					log8.BaseLogger.Info().Msg("RabbitMQ publishing message to exchange for the ASMM8 service.")
					err := amqpM8.WithPooledConnection(func(am8 amqpM8.PooledAmqpInterface) error {
						return am8.Publish(queue[0], "cptm8.asmm8.get.scan", nil, "orchestratorm8")
					})
					if err != nil {
						log8.BaseLogger.Debug().Msg(err.Error())
						log8.BaseLogger.Error().Msgf("RabbitMQ publishing message to exchange for the ASMM8 service - fail")
					} else {
						log8.BaseLogger.Info().Msg("RabbitMQ publishing message to the ASMM8 queue service - success.")
					}
				}
			}
		}
	}

	go checkDBEmpty(true)
	go checkRMQPublish()

	var c chan bool
	<-c
}
