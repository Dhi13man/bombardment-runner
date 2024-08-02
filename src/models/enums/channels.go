package models_enums

type Channel string

const (
	REST  Channel = "REST"
	GRPC  Channel = "GRPC"
	KAFKA Channel = "KAFKA"
)
