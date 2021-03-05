package devices_test

import (
	"testing"

	"github.com/dougedey/elsinore/devices"
	"periph.io/x/periph/conn/onewire"
	"periph.io/x/periph/conn/physic"
)

func TestCreateTemperatureController(t *testing.T) {
	devices.ClearControllers()
	probe := devices.TemperatureProbe{
		PhysAddr: "ARealAddress",
		Address:  onewire.Address(12345),
	}

	t.Run("A new Temperature controller is created if no existing device with the same name exists", func(t *testing.T) {
		temperatureController, err := devices.CreateTemperatureController("sample", &probe)
		if err != nil {
			t.Fatalf("Failed to create the Temperature Controller: %v", err)
		}

		if temperatureController == nil {
			t.Fatalf("No Pid Controller returned for sample")
		}

		if temperatureController.Name != "sample" {
			t.Fatalf("Expected the temperature controller to be called sample, but got %v", temperatureController.Name)
		}
	})

	t.Run("A Temperature controller cannot be created if the probe is already associated with a controller", func(t *testing.T) {
		temperatureController, err := devices.CreateTemperatureController("sample_2", &probe)

		if err == nil {
			t.Fatalf("Created a duplicate Temperature controller: %v", temperatureController)
		}
	})

	t.Run("Re-adding a probe to the same temperature controller is a no-op", func(t *testing.T) {
		temperatureController, err := devices.CreateTemperatureController("sample", &probe)

		if err != nil {
			t.Fatalf("Failed to do nothing: %v", err)
		}

		existingController := devices.FindTemperatureControllerByName("sample")
		if temperatureController != existingController {
			t.Fatalf("Expected %v, but got %v", existingController, temperatureController)
		}
	})
}

func TestTemperatureControllerAverageTemperature(t *testing.T) {
	devices.ClearControllers()
	probe := devices.TemperatureProbe{
		PhysAddr: "ARealAddress",
		Address:  onewire.Address(12345),
		Reading:  physic.Temperature(0),
	}
	temperatureController, err := devices.CreateTemperatureController("sample", &probe)

	if err != nil {
		t.Fatalf("Failed to create the controller: %v", err)
	}

	t.Run("With a single probe, you get the current value", func(t *testing.T) {
		probe.UpdateTemperature("35C")
		avgTemperature := temperatureController.AverageTemperature()
		if float64(35.0) != avgTemperature.Celsius() {
			t.Fatalf("Expected %v, but got %v", probe.Reading, avgTemperature)
		}
	})

	t.Run("With multiple probes, you get an average value", func(t *testing.T) {
		probe_two := devices.TemperatureProbe{
			PhysAddr: "AnotherRealAddress",
			Address:  onewire.Address(123456),
			Reading:  physic.Temperature(0),
		}
		devices.CreateTemperatureController("sample", &probe_two)

		// probe.Reading = new(physic.Temperature)
		// probe.Reading.Set("35C")
		probe_two.UpdateTemperature("37C")

		avgTemperature := temperatureController.AverageTemperature()
		if float64(36.0) != avgTemperature.Celsius() {
			t.Fatalf("Expected 36C, but got %v", avgTemperature)
		}
	})
}
