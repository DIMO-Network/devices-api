package elasticsearch

type UpdateByQueryResponse struct {
	Took     int64                    `json:"took,omitempty"`
	Updated  int64                    `json:"updated,omitempty"`
	Total    int64                    `json:"total,omitempty"`
	Failures []map[string]interface{} `json:"failures,omitempty"`
}
