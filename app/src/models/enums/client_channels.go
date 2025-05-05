package modelsEnums

type ClientChannel string

const (
	REST  ClientChannel = "REST"
	GRPC  ClientChannel = "GRPC"
	KAFKA ClientChannel = "KAFKA"
)
