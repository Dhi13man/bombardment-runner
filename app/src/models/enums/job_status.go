package modelsEnums

type JobStatus string

const (
	PAUSED  JobStatus = "PAUSED"
	RUNNING JobStatus = "RUNNING"
	STOPPED JobStatus = "STOPPED"
)
