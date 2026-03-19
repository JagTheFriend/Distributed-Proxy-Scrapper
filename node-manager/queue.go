package nodemanager

import (
	"common"
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/valkey-io/valkey-glide/go/v2/models"
)

type Geo struct {
	Country string `json:"country"`
	City    string `json:"city"`
}

type Client struct {
	ClientId    string  `json:"clientId"`
	ClientType  string  `json:"clientType"`
	IP          string  `json:"ip"`
	Geo         Geo     `json:"geo"`
	ConnectedAt int64   `json:"connectedAt"`
	LatencyMs   int     `json:"latencyMs"`
	Load        float64 `json:"load"`
}

func AddClient(ctx context.Context, client *Client) error {
	valkey := GetValKeyClient()

	data, _ := json.Marshal(client)

	_, err := valkey.Set(ctx, common.FormatClientKey("available", client.ClientId), string(data))
	return err
}

func GetAvailableClients(ctx context.Context) ([]string, error) {
	cursor := models.NewCursor()
	var clients []string

	for {
		result, err := client.Scan(ctx, cursor)
		if err != nil {
			return nil, err
		}

		for _, key := range result.Data {
			if strings.HasPrefix(key, "client:available:") {
				clients = append(clients, key)
			}
		}

		cursor = result.Cursor
		if cursor.IsFinished() {
			break
		}
	}

	if len(clients) == 0 {
		return nil, errors.New("no available clients")
	}

	return clients, nil
}

func SetClientBusy(ctx context.Context, clientId string) error {
	valkey := GetValKeyClient()
	_, err := valkey.Rename(ctx,
		common.FormatClientKey("available", clientId),
		common.FormatClientKey("occupied", clientId),
	)
	return err
}

func SetClientIdle(ctx context.Context, clientId string) error {
	valkey := GetValKeyClient()
	_, err := valkey.Rename(ctx,
		common.FormatClientKey("occupied", clientId),
		common.FormatClientKey("available", clientId),
	)
	return err
}

func RemoveClient(ctx context.Context, clientId string) error {
	valkey := GetValKeyClient()
	_, err := valkey.Del(ctx, []string{
		common.FormatClientKey("available", clientId),
		common.FormatClientKey("occupied", clientId),
		common.FormatClientKey("blocked", clientId),
	})
	return err
}
