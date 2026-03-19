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

func GetAvailableClient(ctx context.Context) (string, error) {
	cursor := models.NewCursor()

	for {
		result, err := client.Scan(ctx, cursor)
		if err != nil {
			return "", err
		}

		for _, key := range result.Data {

			// Filter only available clients
			if strings.HasPrefix(key, "client:available:") {
				return key, nil
			}
		}

		cursor = result.Cursor
		if cursor.IsFinished() {
			break
		}
	}

	return "", errors.New("No Available Client")
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
