package main

type ValidateRequest struct {
	Text string `json:"text"`
}

type ValidateResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

