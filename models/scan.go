package models

type ScanParamsRequest struct {
	Target string `json:"target"`
	Port   string `json:"port"`
}

type ScanParamsResponse struct {
	Host  string `json:"host"`
	Ports []Port `json:"ports"`
}

type Port struct {
	Port           uint16 `json:"port"`
	Protocol       string `json:"protocol"`
	State          string `json:"state"`
	ServiceName    string `json:"service"`
	ServiceVersion string `json:"version"`
}
