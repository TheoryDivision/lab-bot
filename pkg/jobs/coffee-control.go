package jobs

import (
	"github.com/stianeikeland/go-rpio/v4"
)

func (cj *controllerJob) gpioInit() {
	err := rpio.Open()
	if err != nil {
		cj.logger.WithField("error", err).Error("Cannot access GPIO")

	}

}
