package devices_test

import (
	"testing"

	"github.com/dougedey/elsinore/devices"

	"periph.io/x/periph/conn/onewire"
)

func TestGetTemperature(t *testing.T) {
	t.Run("Invalid addresses return nothing", func(t *testing.T) {
		emptyResult := devices.GetTemperature("NotAnAddress")

		if emptyResult != nil {
			t.Errorf("GetTemperature with an invalid address string failed, expected nil, got %v", emptyResult)
		}
	})

	t.Run("A real address returns the probe", func(t *testing.T) {
		realAddress := "ARealAddress"
		devices.SetProbe(&devices.TemperatureProbe{
			PhysAddr: realAddress,
			Address:  onewire.Address(12345),
		},
		)

		actualProbe := devices.GetTemperature(realAddress)

		if actualProbe == nil {
			t.Errorf("GetTemperature returned a nil object for %v", realAddress)
		}
	})
}

func TestGetAddresses(t *testing.T) {
	t.Run("GetAddresses returns an empty list by default", func(t *testing.T) {
		emptyResult := devices.GetAddresses()

		if len(emptyResult) == 0 {
			t.Errorf("The default list should be empty. Found %v", emptyResult)
		}
	})
}
