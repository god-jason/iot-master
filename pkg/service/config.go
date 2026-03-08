package service

import (
	"github.com/god-jason/iot-master/pkg/config"
)

const MODULE = "service"

func init() {
	config.SetDefault(MODULE, "name", "boat")
	config.SetDefault(MODULE, "display", "Boat")
	config.SetDefault(MODULE, "description", "Process Manager for General IoT Backend")
	config.SetDefault(MODULE, "arguments", []string{})
}
