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

func TestBasicFunctions(t *testing.T) {
	t.Run("A nil temperature probe returns an empty string", func(t *testing.T) {
		p := (*hardware.TemperatureProbe)(nil)
		if p.Reading() != "" {
			t.Errorf("Expected an empty string but got %v", p.Reading())
		}
	})

	p := hardware.TemperatureProbe{PhysAddr: "ATest"}

	t.Run("Update temperature returns errors from the underlying call", func(t *testing.T) {
		err := p.UpdateTemperature(("Foo"))
		if err == nil {
			t.Error("Input of Foo should not be valid")
		}
	})

	t.Run("Update temperature sets the value", func(t *testing.T) {
		err := p.UpdateTemperature(("23C"))
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Reading returns the current temperate", func(t *testing.T) {
		val := p.Reading()
		if val != "23°C" {
			t.Errorf("Temperature should be 23C, but was %v", val)
		}
	})
}
