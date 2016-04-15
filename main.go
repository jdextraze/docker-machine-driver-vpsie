package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/jdextraze/docker-machine-driver-vpsie/driver"
)

func main() {
	plugin.RegisterDriver(driver.NewDriver("", ""))
}
