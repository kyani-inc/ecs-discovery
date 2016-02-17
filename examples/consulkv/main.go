package main

import (
	"fmt"
	"time"

	consul "github.com/hashicorp/consul/api"
	"github.com/kyani-inc/ecs-discovery"
)

func main() {
	kv, err := discovery.ConsulKV(consul.DefaultConfig())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	kv.NewClient("my-test-cluster", "us-east-1", "discovery.mydomain.com")

	for {
		err := kv.Discover()
		if err != nil {
			fmt.Println(err.Error())
		}

		time.Sleep(15 * time.Second) // We recommend you use a backoff
	}
}
