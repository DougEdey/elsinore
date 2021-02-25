package devices

import (
	"testing"

	"periph.io/x/periph/conn/onewire"
)

func TestGetTemperature(t *testing.T) {
	t.Run("Invalid addresses return nothing", func(t *testing.T) {
		emptyResult := GetTemperature("NotAnAddress")

		if emptyResult != nil {
			t.Errorf("GetTemperature with an invalid address string failed, expected nil, got %v", emptyResult)
		}
	})

	t.Run("A real address returns the probe", func(t *testing.T) {
		realAddress := "ARealAddress"
		probes[realAddress] = &TemperatureProbe{
			PhysAddr: realAddress,
			Address:  onewire.Address(12345),
		}

		actualProbe := GetTemperature(realAddress)

		if actualProbe == nil {
			t.Errorf("GetTemperature returned a nil object for %v", realAddress)
		}
	})
}

func TestGetAddresses(t *testing.T) {
	t.Run("GetAddresses returns an empty list by default", func(t *testing.T) {
		emptyResult := GetAddresses()

		if len(emptyResult) == 0 {
			t.Errorf("The default list should be empty. Found %v", emptyResult)
		}
	})
}
