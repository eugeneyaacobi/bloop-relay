package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type enrollRequest struct {
	Token         string `json:"token"`
	ClientName    string `json:"clientName"`
	ClientVersion string `json:"clientVersion"`
}

type enrollResponse struct {
	Installation struct {
		ID string `json:"id"`
	} `json:"installation"`
	Ingest struct {
		Token string `json:"token"`
	} `json:"ingest"`
}

func EnrollRuntime(ctx context.Context, controlPlaneURL, enrollmentToken, clientName string) (string, string, error) {
	controlPlaneURL = strings.TrimRight(strings.TrimSpace(controlPlaneURL), "/")
	if controlPlaneURL == "" || strings.TrimSpace(enrollmentToken) == "" {
		return "", "", fmt.Errorf("missing control plane url or enrollment token")
	}
	payload, _ := json.Marshal(enrollRequest{Token: enrollmentToken, ClientName: clientName, ClientVersion: "0.1.0"})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, controlPlaneURL+"/api/runtime/enroll", bytes.NewReader(payload))
	if err != nil { return "", "", err }
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil { return "", "", err }
	defer resp.Body.Close()
	if resp.StatusCode >= 300 { return "", "", fmt.Errorf("runtime enroll status %d", resp.StatusCode) }
	var parsed enrollResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil { return "", "", err }
	if parsed.Installation.ID == "" || parsed.Ingest.Token == "" {
		return "", "", fmt.Errorf("runtime enroll response missing installation id or ingest token")
	}
	return parsed.Installation.ID, parsed.Ingest.Token, nil
}
