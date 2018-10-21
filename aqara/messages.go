package aqara

// ListenResponse is a map of Gateway -> Message
type ListenResponse struct {
	Message ReportMessage
	Gateway *Gateway
}

// ReportMessage are returned from Listen
type ReportMessage struct {
	Model string
	Sid   string
	Data  map[string]interface{}
}

type internalReportMessage struct {
	// {"cmd":"report","model":"magnet","sid":"89234324","short_id":4343,"data":"{\"status\":\"open\"}"}
	Cmd   string `json:"cmd"`
	Model string `json:"model"`
	Sid   string `json:"sid"`
	Token string `json:"token,omitempty"`
	Data  string `json:"data"`
}
