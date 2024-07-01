package main

import (
	// TODO: update the path below to match your own repository

	"context"
	"fmt"
	"github/juanadelacuesta/heraclitus/libvirt"
	"github/juanadelacuesta/heraclitus/virt"
	"time"

	"github.com/hashicorp/go-hclog"
)

func main() {

	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:  "my-app",
		Level: hclog.LevelFromString("DEBUG"),
	})
	ctx, cancel := context.WithCancel(context.Background())

	conn, err := libvirt.New(ctx, "qemu:///system", appLogger)
	if err != nil {
		fmt.Printf("error: %+v\n %+v\n", conn, err)
		return
	}

	//conn.GetVms()

	config := &libvirt.DomainConfig{
		Name:     "blah",
		Metadata: map[string]string{"ID": "blah"},
		Memory:   10,
		CPUs:     1,
		Cores:    2,
	}

	conn.CreateDomain(config)
	//conn.GetVms()
	cancel()

	time.Sleep(2 * time.Second)

	// Serve the plugin
	//plugins.Serve(factory)
}

// factory returns a new instance of a nomad driver plugin
func factory(log hclog.Logger) interface{} {
	return virt.NewPlugin(log)
}
