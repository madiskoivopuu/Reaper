package rblx

import "net/http"

// Main struct for roblox sessions
type RBLXSession struct {
	Cookie     string
	Username   string
	UserID     int64
	XCSRFToken string
	Client     *http.Client
}
