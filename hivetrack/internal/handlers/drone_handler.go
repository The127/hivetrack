package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/the127/hivetrack/internal/config"
	"github.com/the127/hivetrack/internal/models"
)

// DroneHandler proxies drone management requests to Hivemind's HTTP management API.
type DroneHandler struct {
	managementURL   string
	managementToken string
	client          *http.Client
}

func NewDroneHandler(cfg *config.HivemindConfig) *DroneHandler {
	return &DroneHandler{
		managementURL:   cfg.ManagementURL,
		managementToken: cfg.ManagementToken,
		client:          &http.Client{},
	}
}

// ListDrones proxies GET /projects/{slug}/drones → GET hivemind/api/v1/drones?project_slug={slug}
func (h *DroneHandler) ListDrones(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]
	h.proxyTo(w, "GET", fmt.Sprintf("/api/v1/drones?project_slug=%s", slug), nil)
}

// CreateToken proxies POST /projects/{slug}/drones/tokens → POST hivemind/api/v1/drones/tokens
func (h *DroneHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}
	body["project_slug"] = slug

	data, _ := json.Marshal(body)
	h.proxyTo(w, "POST", "/api/v1/drones/tokens", bytes.NewReader(data))
}

// GetDrone proxies GET /projects/{slug}/drones/{drone_id} → GET hivemind/api/v1/drones/{drone_id}
func (h *DroneHandler) GetDrone(w http.ResponseWriter, r *http.Request) {
	droneID := mux.Vars(r)["drone_id"]
	h.proxyTo(w, "GET", fmt.Sprintf("/api/v1/drones/%s", droneID), nil)
}

// DeregisterDrone proxies POST /projects/{slug}/drones/{drone_id}/deregister
func (h *DroneHandler) DeregisterDrone(w http.ResponseWriter, r *http.Request) {
	droneID := mux.Vars(r)["drone_id"]
	h.proxyTo(w, "POST", fmt.Sprintf("/api/v1/drones/%s/deregister", droneID), nil)
}

// RevokeToken proxies DELETE /projects/{slug}/drones/tokens/{token}
func (h *DroneHandler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	h.proxyTo(w, "DELETE", fmt.Sprintf("/api/v1/drones/tokens/%s", token), nil)
}

func (h *DroneHandler) proxyTo(w http.ResponseWriter, method, path string, body io.Reader) {
	url := h.managementURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		logger.Error("creating hivemind proxy request", zap.Error(err))
		RespondError(w, fmt.Errorf("proxy error: %w", err))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if h.managementToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.managementToken)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		logger.Error("proxying to hivemind", zap.Error(err))
		RespondError(w, fmt.Errorf("hivemind unavailable: %w", err))
		return
	}
	defer func() { _ = resp.Body.Close() }()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
