// Package models contains the request, response, and domain types for the guide service.
package models

import "time"

// EmergencyGuide is a single piece of emergency guidance content.
type EmergencyGuide struct {
	ID               string    `json:"id"`
	HazardType       string    `json:"hazardType"`
	Stage            string    `json:"stage"`
	Title            string    `json:"title"`
	Body             string    `json:"body"`
	Language         string    `json:"language"`
	OfflineAvailable bool      `json:"offlineAvailable"`
	SortOrder        int       `json:"sortOrder"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// GuideListResponse is the payload returned when listing guides.
type GuideListResponse struct {
	Guides []EmergencyGuide `json:"guides"`
}

// GuideFilters captures the accepted query parameters for listing guides.
type GuideFilters struct {
	HazardType string
	Stage      string
	Language   string
	Offline    *bool
}

// APIError is the standard error response envelope.
type APIError struct {
	Error APIErrorBody `json:"error"`
}

// APIErrorBody is the standard error response body.
type APIErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
