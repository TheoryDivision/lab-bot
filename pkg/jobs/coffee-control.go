package jobs

import (
	"github.com/stianeikeland/go-rpio/v4"
)

const pin = 17

func pinInit() (err error) {
	err = rpio.Open()
	if err != nil {
		return err
	}
	p := rpio.Pin(pin)
	p.Output()
	p.Low()
	return nil
}

func pinOn() (err error) {
	p := rpio.Pin(pin)
	p.Output()
	p.High()
	return nil
}

func pinOff() (err error) {
	p := rpio.Pin(pin)
	p.Output()
	p.Low()
	return nil
}
