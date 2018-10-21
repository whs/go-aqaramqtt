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
	Data  DeviceData
}

type internalReportMessage struct {
	// {"cmd":"report","model":"magnet","sid":"89234324","short_id":4343,"data":"{\"status\":\"open\"}"}
	Cmd   string `json:"cmd"`
	Model string `json:"model"`
	Sid   string `json:"sid"`
	Token string `json:"token,omitempty"`
	Data  string `json:"data"`
}

// DeviceData are reported from ReportMessage
type DeviceData struct {
	Status       string  `json:"status,omitempty"`
	Battery      int     `json:"battery,omitempty,string"`
	Voltage      int     `json:"voltage,omitempty"`
	LoadPower    float32 `json:"load_power,omitempty,string"`
	Error        string  `json:"error,omitempty"`
	Illumination int     `json:"illumination,omitempty"`
}
