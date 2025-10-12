package main

import (
	"net/http"

	"github.com/google/uuid"
)

// generateGUID creates a new GUID using the google/uuid package
func generateGUID() string {
	return uuid.New().String()
}

// getUserID checks for existing userId cookie or creates a new one
func getUserID(w http.ResponseWriter, r *http.Request) (string, error) {
	// Check for existing cookie
	cookie, err := r.Cookie("chat-userid")
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	// Generate new GUID
	userID := generateGUID()

	// Set the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "chat-userid",
		Value:    userID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return userID, nil
}
