package controller8

import (
	"context"
	"database/sql"
	amqpM8 "deifzar/orchestratorm8/pkg/amqpM8"
	"deifzar/orchestratorm8/pkg/cleanup8"
	"deifzar/orchestratorm8/pkg/db8"
	"deifzar/orchestratorm8/pkg/log8"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/spf13/viper"

	_ "github.com/lib/pq"
)

type Controller8OrchestratorM8 struct {
	Config *viper.Viper
	Db     *sql.DB
}

func NewController8OrchestratorM8(db *sql.DB, cnfg *viper.Viper) (Controller8OrchestratorM8Interface, error) {

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
		err := amqpM8.InitializeConnectionPool()
		if err != nil {
			log8.BaseLogger.Debug().Msg(err.Error())
			return &Controller8OrchestratorM8{}, err
		}
	}

	o := &Controller8OrchestratorM8{Db: db, Config: cnfg}
	return o, nil
}

func (o *Controller8OrchestratorM8) InitOrchestrator() error {
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

func (o *Controller8OrchestratorM8) StartOrchestrator(ctx context.Context) {

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
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log8.BaseLogger.Info().Msg("checkDBEmpty goroutine shutting down")
				return
			case <-ticker.C:
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
			}
		}
	}
	checkRMQPublish := func() {
		for {
			var empty, publish bool

			select {
			case <-ctx.Done():
				log8.BaseLogger.Info().Msg("checkRMQPublish goroutine shutting down")
				return
			case empty = <-emptyChan:
				publish = <-publishChan
			}

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

	<-ctx.Done()

	// Cleanup code here (close DB connections, etc.)
	log8.BaseLogger.Info().Msg("Orchestrator stopped")
}

func (o *Controller8OrchestratorM8) ExistQueue(queueName string, queueArgs amqp.Table) bool {
	var exists bool

	err := amqpM8.WithPooledConnection(func(am8 amqpM8.PooledAmqpInterface) error {
		exists = am8.ExistQueue(queueName, queueArgs)
		return nil
	})

	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msgf("Error checking if queue `%s` exists", queueName)
		return false
	}

	return exists
}

func (o *Controller8OrchestratorM8) ExistConsumersForQueue(queueName string) bool {
	var exists bool

	err := amqpM8.WithPooledConnection(func(am8 amqpM8.PooledAmqpInterface) error {
		c := am8.GetConsumersForQueue(queueName)
		exists = len(c) > 0
		return nil
	})

	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msgf("Error checking if consumers exist for queue `%s`", queueName)
		return false
	}

	return exists
}

func (m *Controller8OrchestratorM8) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "orchestratorm8",
	})
}

func (m *Controller8OrchestratorM8) ReadinessCheck(c *gin.Context) {
	dbHealthy := true
	rbHealthy := true
	if err := m.Db.Ping(); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Database ping failed during readiness check")
		dbHealthy = false
	}
	dbStatus := "unhealthy"
	if dbHealthy {
		dbStatus = "healthy"
	}

	queue_consumer := m.Config.GetStringSlice("ORCHESTRATORM8.asmm8.Queue")
	qargs_consumer := m.Config.GetStringMap("ORCHESTRATORM8.asmm8.Queue-arguments")

	if !m.ExistQueue(queue_consumer[1], qargs_consumer) || !m.ExistConsumersForQueue(queue_consumer[1]) {
		rbHealthy = false
	}

	rbStatus := "unhealthy"
	if rbHealthy {
		rbStatus = "healthy"
	}

	if dbHealthy && rbHealthy {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "orchestratorm8",
			"checks": gin.H{
				"database": dbStatus,
				"rabbitmq": rbStatus,
			},
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not ready",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "orchestratorm8",
			"checks": gin.H{
				"database": dbStatus,
				"rabbitmq": rbStatus,
			},
		})
	}
}
