package modelsEnums

type ClientChannel string

const (
	REST    ClientChannel = "REST"
	GRAPHQL ClientChannel = "GRAPHQL"
	GRPC    ClientChannel = "GRPC"
	KAFKA   ClientChannel = "KAFKA"
)
