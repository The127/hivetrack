package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetHivemindConfig returns the Hivemind gRPC URL for drone setup.
func (c *Client) GetHivemindConfig(ctx context.Context) (*HivemindConfig, error) {
	data, err := c.get(ctx, "/api/v1/hivemind/config", nil)
	if err != nil {
		return nil, err
	}
	var cfg HivemindConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

// ListDrones returns all drones registered for a project.
func (c *Client) ListDrones(ctx context.Context, slug string) ([]Drone, error) {
	data, err := c.get(ctx, "/api/v1/projects/"+slug+"/drones", nil)
	if err != nil {
		return nil, err
	}
	var drones []Drone
	if err := json.Unmarshal(data, &drones); err != nil {
		return nil, fmt.Errorf("parsing drones: %w", err)
	}
	return drones, nil
}

// GetDrone returns a single drone by ID.
func (c *Client) GetDrone(ctx context.Context, slug, droneID string) (*Drone, error) {
	data, err := c.get(ctx, fmt.Sprintf("/api/v1/projects/%s/drones/%s", slug, droneID), nil)
	if err != nil {
		return nil, err
	}
	var drone Drone
	if err := json.Unmarshal(data, &drone); err != nil {
		return nil, fmt.Errorf("parsing drone: %w", err)
	}
	return &drone, nil
}

// CreateDroneToken creates a registration token for a new drone in a project.
func (c *Client) CreateDroneToken(ctx context.Context, slug string, req CreateDroneTokenRequest) (*CreateDroneTokenResult, error) {
	data, err := c.post(ctx, "/api/v1/projects/"+slug+"/drones/tokens", req)
	if err != nil {
		return nil, err
	}
	var result CreateDroneTokenResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing result: %w", err)
	}
	return &result, nil
}

// DeregisterDrone deregisters a drone from Hivemind without deleting its record.
func (c *Client) DeregisterDrone(ctx context.Context, slug, droneID string) error {
	_, err := c.post(ctx, fmt.Sprintf("/api/v1/projects/%s/drones/%s/deregister", slug, droneID), nil)
	return err
}

// DeleteDrone permanently deletes a drone.
func (c *Client) DeleteDrone(ctx context.Context, slug, droneID string) error {
	_, err := c.delete(ctx, fmt.Sprintf("/api/v1/projects/%s/drones/%s", slug, droneID))
	return err
}

// RevokeDroneToken revokes a drone registration token.
func (c *Client) RevokeDroneToken(ctx context.Context, slug, token string) error {
	_, err := c.delete(ctx, fmt.Sprintf("/api/v1/projects/%s/drones/tokens/%s", slug, token))
	return err
}
