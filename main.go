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
	ctx, cancel := context.WithCancel(context.Background())
	conn, err := libvirt.New(ctx, "qemu:///system", hclog.Default().Named("blah"))
	fmt.Printf("\n %+v\n %+v", conn, err)
	// Serve the plugin
	//plugins.Serve(factory)
	conn.GetVms()
	cancel()
	time.Sleep(2 * time.Second)
}

// factory returns a new instance of a nomad driver plugin
func factory(log hclog.Logger) interface{} {
	return virt.NewPlugin(log)
}
