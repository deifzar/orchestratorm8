package orchestratorm8

type Orchestrator8Interface interface {
	InitOrchestrator() error
	StartOrchestrator()
}
