package nodemanager

import (
	"common"
	"fmt"
	"strconv"

	glide "github.com/valkey-io/valkey-glide/go/v2"
	"github.com/valkey-io/valkey-glide/go/v2/config"
)

var client *glide.Client

func GetValKeyClient() *glide.Client {
	if client != nil {
		return client
	}

	host, err := common.GetEnv("VALKEY_HOST")
	if err != nil {
		panic("VALKEY_HOST not set")
	}

	port, err := common.GetEnv("VALKEY_PORT")
	if err != nil {
		panic("VALKEY_PORT not set")
	}

	parsedPort, err := strconv.Atoi(port)
	if err != nil {
		panic("Error parsing VALKEY_PORT")
	}

	config := config.NewClientConfiguration().
		WithAddress(&config.NodeAddress{Host: host, Port: parsedPort})

	newClient, err := glide.NewClient(config)
	if err != nil {
		panic(fmt.Sprintf("Error connecting to valkey: %s", err.Error()))
	}

	client = newClient
	return newClient
}
