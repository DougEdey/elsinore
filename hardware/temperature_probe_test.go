package hardware_test

import (
	"testing"

	"github.com/dougedey/elsinore/hardware"

	"periph.io/x/periph/conn/onewire"
)

func TestGetTemperature(t *testing.T) {
	t.Run("Invalid addresses return nothing", func(t *testing.T) {
		emptyResult := hardware.GetTemperature("NotAnAddress")

		if emptyResult != nil {
			t.Errorf("GetTemperature with an invalid address string failed, expected nil, got %v", emptyResult)
		}
	})

	t.Run("A real address returns the probe", func(t *testing.T) {
		realAddress := "ARealAddress"
		hardware.SetProbe(&hardware.TemperatureProbe{
			PhysAddr: realAddress,
			Address:  onewire.Address(12345),
		},
		)

		actualProbe := hardware.GetTemperature(realAddress)

		if actualProbe == nil {
			t.Errorf("GetTemperature returned a nil object for %v", realAddress)
		}
	})
}

func TestGetProbes(t *testing.T) {
	t.Run("GetProbes returns an empty list by default", func(t *testing.T) {
		emptyResult := hardware.GetProbes()

		if len(emptyResult) == 0 {
			t.Errorf("The default list should be empty. Found %v", emptyResult)
		}
	})
}
