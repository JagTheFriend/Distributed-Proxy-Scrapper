package nodemanager

import (
	"context"
	"encoding/json"
)

type Geo struct {
	Country string `json:"country"`
	City    string `json:"city"`
}

type Client struct {
	ClientId    string  `json:"clientId"`
	ClientType  string  `json:"clientType"` // web, mobile etc
	IP          string  `json:"ip"`
	Geo         Geo     `json:"geo"`
	ConnectedAt int64   `json:"connectedAt"` // Unix timestamp
	LatencyMs   int     `json:"latencyMs"`
	Load        float64 `json:"load"`
}

func GetClient(ctx context.Context, clientId string) (*Client, error) {
	valkey := GetValKeyClient()

	result, err := valkey.Get(ctx, "client:available:"+clientId)
	if err != nil {
		return nil, err
	}

	var client Client
	if err := json.Unmarshal([]byte(result.Value()), &client); err != nil {
		return nil, err
	}

	return &client, nil
}

func AddClient(ctx context.Context, client *Client) error {
	valkey := GetValKeyClient()

	// Convert struct to JSON
	data, err := json.Marshal(client)
	if err != nil {
		return err
	}

	_, err = valkey.Set(ctx, "client:available:"+client.ClientId, string(data))
	return err
}

func RemoveClient(ctx context.Context, clientId string) error {
	valkey := GetValKeyClient()
	_, err := valkey.Del(ctx, []string{"client:available:" + clientId, "client:occupied:" + clientId, "client:blocked:" + clientId})
	return err
}

func AddClientToOccupied(ctx context.Context, clientId string) error {
	valkey := GetValKeyClient()
	_, err := valkey.Rename(ctx, "client:available:"+clientId, "client:occupied:"+clientId)
	return err
}

func AddClientToBlocked(ctx context.Context, clientId string) error {
	valkey := GetValKeyClient()
	_, err := valkey.Rename(ctx, "client:occupied:"+clientId, "client:blocked:"+clientId)
	return err
}
