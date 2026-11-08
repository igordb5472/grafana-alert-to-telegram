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

	errCh := make(chan error)
	var body []byte
	go func() {
		var err error
		body, err = io.ReadAll(req.Body)
		if ctx.Err() == nil {
			errCh <- err
		}
	}()
	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "request timeout exceeded while reading request body", http.StatusGatewayTimeout)
		}
		return
	case err := <-errCh:
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read request body: %v", err), http.StatusInternalServerError)
			return
		}
	}

	if len(body) == 0 {
		http.Error(w, "request body must be non-empty", http.StatusBadRequest)
		return
	}
	data := make(map[string]any)
	go func() {
		err := json.Unmarshal(body, &data)
		if ctx.Err() == nil {
			errCh <- err
		}
	}()
	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "request timeout exceeded while unmarshalling request body data", http.StatusGatewayTimeout)
		}
		return
	case err := <-errCh:
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to unmarshall request body data: %v", err), http.StatusInternalServerError)
			return
		}
	}

	go func() {
		err := service.SendAllAlertNotifs(ctx, data)
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "request timeout exceeded while unmarshalling request body data", http.StatusGatewayTimeout)
			return
		}
		return
	case err := <-errCh:
		if err != nil {
			http.Error(w, fmt.Sprintf("HandleAlert: %v", err), http.StatusInternalServerError)
			return
		}
	}
}
