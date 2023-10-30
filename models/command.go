package models

type CommandRequest struct {
	Command string `json:"command"`
}

type CommandResponse struct {
	Result string `json:"result"`
}