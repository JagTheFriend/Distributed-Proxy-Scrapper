package nodemanager

import (
	"context"
	"encoding/json"
)

type Client struct {
	ClientId string `json:"clientId"`
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

func AddClient(ctx context.Context, clientId string) error {
	valkey := GetValKeyClient()

	// Convert struct to JSON
	data, err := json.Marshal(clientId)
	if err != nil {
		return err
	}

	_, err = valkey.Set(ctx, "client:available:"+clientId, string(data))
	return err
}

func RemoveClient(ctx context.Context, clientId string) error {
	valkey := GetValKeyClient()
	_, err := valkey.Del(ctx, []string{"client:" + clientId})
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
