package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"grafana-alert-to-telegram/src/service"
	"io"
	"net/http"
)

func HandleAlert(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if ctx.Err() != nil {
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "request timeout exceeded while reading request body from Grafana", http.StatusGatewayTimeout)
		}
		return
	}

	username, password, ok := req.BasicAuth()
	if !ok {
		http.Error(w, "basic authorization header must be configured", http.StatusUnauthorized)
		return
	}

	if !service.AuthOK(username, password) {
		http.Error(w, "wrong credentials in basic authorization header", http.StatusForbidden)
		return
	}

	contentType, ok := req.Header["Content-Type"]
	if !ok {
		http.Error(w, "request Content-Type must be configured", http.StatusBadRequest)
		return
	}
	if len(contentType) != 1 || contentType[0] != "application/json" {
		http.Error(w, "request Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	userAgent, ok := req.Header["User-Agent"]
	if !ok {
		http.Error(w, "request User-Agent must be configured", http.StatusBadRequest)
		return
	}
	if len(userAgent) != 1 || userAgent[0] != "Grafana" {
		http.Error(w, "request User-Agent must be Grafana", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if ctx.Err() != nil {
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "request timeout exceeded while reading request body from Grafana", http.StatusGatewayTimeout)
		}
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read request body from Grafana: %v", err), http.StatusInternalServerError)
		return
	}
	if len(body) == 0 {
		http.Error(w, "request body from Grafana must be non-empty", http.StatusBadRequest)
		return
	}

	data := make(map[string]any)
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to unmarshall request body from Grafana: %v", err), http.StatusInternalServerError)
		return
	}

	err = service.SendAllAlertNotifs(ctx, data)
	if ctx.Err() != nil {
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "request timeout exceeded while unmarshalling request body from Grafana", http.StatusGatewayTimeout)
		}
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("HandleAlert: %v", err), http.StatusInternalServerError)
		return
	}
}
