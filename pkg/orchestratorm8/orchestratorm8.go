package orchestratorm8

import (
	"database/sql"
	amqpM8 "deifzar/orchestratorm8/pkg/amqpM8"
	"deifzar/orchestratorm8/pkg/configparser"
	"deifzar/orchestratorm8/pkg/db8"
	"deifzar/orchestratorm8/pkg/log8"
	"time"

	"github.com/spf13/viper"

	_ "github.com/lib/pq"
)

type Orchestrator8 struct {
	Config *viper.Viper
	Amqp   amqpM8.AmqpM8Interface
	Db     *sql.DB
}

func NewOrchestratorM8() (Orchestrator8Interface, error) {

	v, err := configparser.InitConfigParser()
	if err != nil {
		return &Orchestrator8{}, err
	}

	locationRMQ := v.GetString("RabbitMQ.location")
	portRMQ := v.GetInt("RabbitMQ.port")
	usernameRMQ := v.GetString("RabbitMQ.username")
	passwordRMQ := v.GetString("RabbitMQ.password")

	locationDB := v.GetString("Database.location")
	portDB := v.GetInt("Database.port")
	schemaDB := v.GetString("Database.schema")
	databaseDB := v.GetString("Database.database")
	usernameDB := v.GetString("Database.username")
	passwordDB := v.GetString("Database.password")

	am8, err := amqpM8.NewAmqpM8(locationRMQ, portRMQ, usernameRMQ, passwordRMQ)
	// defer am8.CloseConnection()
	// defer am8.CloseChannel()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msg("Error connecting into RabbitMQ.")
		return &Orchestrator8{}, err
	}

	var db db8.Db8
	db.InitDatabase8(locationDB, portDB, schemaDB, databaseDB, usernameDB, passwordDB)
	conn, err := db.OpenConnection()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msg("Error connecting into DB.")
		return &Orchestrator8{}, err
	}

	o := &Orchestrator8{Config: v, Amqp: am8, Db: conn}
	return o, nil
}

func (o *Orchestrator8) InitOrchestrator() error {
	exchanges := o.Config.GetStringMapString("ORCHESTRATORM8.Exchanges")

	for exname, extype := range exchanges {
		err := o.Amqp.DeclareExchange(exname, extype)
		if err != nil {
			log8.BaseLogger.Debug().Msg(err.Error())
			return err
		}
	}
	// Declare 'ASMM8' queue and bind it to the 'CPTM8' Exchange
	queue := o.Config.GetStringSlice("ORCHESTRATORM8.asmm8.Queue")
	qargs := o.Config.GetStringMap("ORCHESTRATORM8.asmm8.Queue-arguments")
	amqp8 := o.Amqp
	log8.BaseLogger.Info().Msg("RabbitMQ declaring queues for the ASMM8 service.")
	err := amqp8.DeclareQueueAndBind(queue[0], queue[1], queue[2], 0, qargs)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return err
	}

	// Declare 'REPORTINGM8' queue and bind it to the 'SCHEDULER' Exchange
	queue = o.Config.GetStringSlice("ORCHESTRATORM8.reportingm8.Queue")
	qargs = o.Config.GetStringMap("ORCHESTRATORM8.reportingm8.Queue-arguments")
	log8.BaseLogger.Info().Msg("RabbitMQ declaring queues for the REPORTINGM8 service.")
	err = amqp8.DeclareQueueAndBind(queue[0], queue[1], queue[2], 0, qargs)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return err
	}

	return nil
}

func (o *Orchestrator8) StartOrchestrator() {

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
					// defer o.Amqp.CloseConnection()
					// defer o.Amqp.CloseChannel()
					queue := o.Config.GetStringSlice("ORCHESTRATORM8.asmm8.Queue")
					log8.BaseLogger.Info().Msg("RabbitMQ publishing message to queue for the ASMM8 service.")
					err := o.Amqp.Publish(queue[0], "cptm8.asmm8.get.scan", "")
					if err != nil {
						log8.BaseLogger.Debug().Msg(err.Error())
						log8.BaseLogger.Error().Msgf("RabbitMQ publishing message to queue for ASMM8 service - fail")
					} else {
						log8.BaseLogger.Info().Msg("RabbitMQ publishing message to ASMM8 queue service - success.")
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
