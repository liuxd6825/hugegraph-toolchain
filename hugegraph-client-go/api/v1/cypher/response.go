package cypher

type ResponseData struct {
	RequestID string `json:"requestId,omitempty"`
	Status    struct {
		Message    string `json:"message"`
		Code       int    `json:"code"`
		Attributes struct {
		} `json:"attributes"`
	} `json:"status"`
	Result struct {
		Data interface{} `json:"data"`
		Meta interface{} `json:"meta"`
	} `json:"result,omitempty"`
	Exception string   `json:"exception,omitempty"`
	Message   string   `json:"message,omitempty"`
	Cause     string   `json:"cause,omitempty"`
	Trace     []string `json:"trace,omitempty"`
}

type ResponseDataResult struct {
	Data interface{} `json:"data"`
	Meta interface{} `json:"meta"`
}

type ResponseDataStatus struct {
	Message    string `json:"message"`
	Code       int    `json:"code"`
	Attributes struct {
	} `json:"attributes"`
}
