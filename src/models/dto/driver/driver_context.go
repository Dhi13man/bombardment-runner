package models_dto_driver

type DriverContext struct {
	BatchSize            int  `json:"batch_size"`
	ShouldStoreResponses bool `json:"should_store_responses"`
}
