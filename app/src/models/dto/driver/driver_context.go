package modelsDtoDriver

type DriverContext struct {
	BatchSize            int    `json:"batch_size"`
	ShouldStoreResponses bool   `json:"should_store_responses"`
	ResponsesStoragePath string `json:"responses_storage_path"`
}
