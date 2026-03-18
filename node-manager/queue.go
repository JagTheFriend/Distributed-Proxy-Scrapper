package nodemanager

import (
	"context"
	"encoding/json"
)

type Client struct {
	ClientId   string  `json:"clientId"`
	Status     int     `json:"status"` // 0 = Available, 1 = Occupied, 2 = Blocked
	ClientType *string `json:"clientType"`
}

func GetClient(ctx context.Context, clientId string) (*Client, error) {
	valkey := GetValKeyClient()

	result, err := valkey.Get(ctx, "client:"+clientId)
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

	_, err = valkey.Set(ctx, "client:"+client.ClientId, string(data))
	return err
}

func RemoveClient(ctx context.Context, clientId string) error {
	valkey := GetValKeyClient()
	_, err := valkey.Del(ctx, []string{"client:" + clientId})
	return err
}
