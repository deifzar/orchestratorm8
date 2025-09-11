package controller8

import (
	"context"

	"github.com/gin-gonic/gin"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Controller8OrchestratorM8Interface interface {
	// Declare RabbmitMQ exchanges and queues for all the services
	InitOrchestrator() error
	// Start two go routines:
	// One that waits for an entry in the 'domain' DB table everytime the table gets empty.
	// Another one that publishes a message to RabbitMQ for ASMM8 to kick off again when the 'domain' DB table is not empty any more.
	StartOrchestrator(ctx context.Context)
	ExistQueue(queueName string, queueArgs amqp.Table) bool
	ExistConsumersForQueue(queueName string, queueArgs amqp.Table) bool
	HealthCheck(c *gin.Context)
	ReadinessCheck(c *gin.Context)
}
