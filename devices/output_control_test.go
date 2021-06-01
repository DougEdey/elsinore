package devices_test

import (
	"log"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/dougedey/elsinore/devices"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiotest"
)

func TestOutputControl(t *testing.T) {
	jumpDuration, err := time.ParseDuration("2.1s")
	if err != nil {
		log.Fatal(err)
	}

	heatPin := gpiotest.Pin{N: "GPIO21", Num: 10, Fn: "I2C1_SDA"}
	outHeat := devices.OutPin{Identifier: "GPIO21", FriendlyName: "GPIO21", PinIO: &heatPin}
	coolPin := gpiotest.Pin{N: "GPIO20", Num: 11, Fn: "I2C1_SDC"}
	outCool := devices.OutPin{Identifier: "GPIO22", FriendlyName: "GPIO22", PinIO: &coolPin}
	outputControl := devices.OutputControl{HeatOutput: &outHeat, CoolOutput: &outCool, DutyCycle: 50, CycleTime: 4}
	outputControl.Reset()

	t.Run("OutputControl turns off imediately when there's no prior state", func(t *testing.T) {
		if l := heatPin.Read(); l != gpio.Low {
			t.Fatal("expected off")
		}

		outputControl.CalculateOutput()

		if l := heatPin.Read(); l != gpio.Low {
			t.Fatal("expected off")
		}
	})

	t.Run("OutputControl waits for the off cycle time before turning on for 2s with a 50% duty cycle on a time of 4s", func(t *testing.T) {
		if l := heatPin.Read(); l != gpio.Low {
			t.Fatal("expected off")
		}

		patch := monkey.Patch(time.Since, func(time.Time) time.Duration { return jumpDuration })
		defer patch.Unpatch()

		outputControl.CalculateOutput()

		if l := heatPin.Read(); l != gpio.High {
			t.Fatal("expected on")
		}
	})

	t.Run("OutputControl waits for the on cycle time before turning off for 2s with a 50% duty cycle on a time of 4s", func(t *testing.T) {
		if l := heatPin.Read(); l != gpio.High {
			t.Fatal("expected on")
		}

		patch := monkey.Patch(time.Since, func(time.Time) time.Duration { return jumpDuration })
		defer patch.Unpatch()

		outputControl.CalculateOutput()

		if l := heatPin.Read(); l != gpio.Low {
			t.Fatal("expected off")
		}
	})

	t.Run("When duty cycle is 100, the cool output is off and the heat output is on", func(t *testing.T) {
		if l := heatPin.Read(); l != gpio.Low {
			t.Fatal("expected off")
		}

		patch := monkey.Patch(time.Since, func(time.Time) time.Duration { return jumpDuration })
		defer patch.Unpatch()

		outputControl.DutyCycle = 100
		outputControl.CalculateOutput()

		if l := heatPin.Read(); l != gpio.High {
			t.Fatal("expected on")
		}
	})

	t.Run("When duty cycle is -100, the heat output is off and the cool output is on", func(t *testing.T) {
		if l := heatPin.Read(); l != gpio.High {
			t.Fatal("expected heat on")
		}

		if l := coolPin.Read(); l != gpio.Low {
			t.Fatal("expected cool off")
		}

		patch := monkey.Patch(time.Since, func(time.Time) time.Duration { return jumpDuration })
		defer patch.Unpatch()

		outputControl.DutyCycle = -100
		outputControl.CalculateOutput()

		if l := heatPin.Read(); l != gpio.Low {
			t.Fatal("expected heat off")
		}
		if l := coolPin.Read(); l != gpio.High {
			t.Fatal("expected cool on")
		}
	})

	t.Run("When duty cycle is 0, heating output is switched off", func(t *testing.T) {
		outHeat.PinIO.Out(gpio.High)
		outCool.PinIO.Out(gpio.Low)
		if l := heatPin.Read(); l != gpio.High {
			t.Fatal("expected heat on")
		}

		if l := coolPin.Read(); l != gpio.Low {
			t.Fatal("expected cool off")
		}

		patch := monkey.Patch(time.Since, func(time.Time) time.Duration { return jumpDuration })
		defer patch.Unpatch()

		outputControl.DutyCycle = 0
		outputControl.CalculateOutput()

		if l := heatPin.Read(); l != gpio.Low {
			t.Fatal("expected heat off")
		}
		if l := coolPin.Read(); l != gpio.Low {
			t.Fatal("expected cool off")
		}
	})

	t.Run("When duty cycle is 0, Cooling output is switched off", func(t *testing.T) {
		outHeat.PinIO.Out(gpio.Low)
		outCool.PinIO.Out(gpio.High)
		if l := heatPin.Read(); l != gpio.Low {
			t.Fatal("expected heat off")
		}

		if l := coolPin.Read(); l != gpio.High {
			t.Fatal("expected cool on")
		}

		patch := monkey.Patch(time.Since, func(time.Time) time.Duration { return jumpDuration })
		defer patch.Unpatch()

		outputControl.DutyCycle = 0
		outputControl.CalculateOutput()

		if l := heatPin.Read(); l != gpio.Low {
			t.Fatal("expected heat off")
		}
		if l := coolPin.Read(); l != gpio.Low {
			t.Fatal("expected cool off")
		}
	})

	t.Run("OutputControl waits for the off cycle time before turning cooling on for 2s with a 50% duty cycle on a time of 4s", func(t *testing.T) {
		if l := coolPin.Read(); l != gpio.Low {
			t.Fatal("expected cooling off")
		}

		patch := monkey.Patch(time.Since, func(time.Time) time.Duration { return jumpDuration })
		defer patch.Unpatch()

		outputControl.DutyCycle = -50
		outputControl.CalculateOutput()

		if l := coolPin.Read(); l != gpio.High {
			t.Fatal("expected cooling on")
		}
	})

	t.Run("OutputControl waits for the on cycle time before turning cooling off for 2s with a 50% duty cycle on a time of 4s", func(t *testing.T) {
		if l := coolPin.Read(); l != gpio.High {
			t.Fatal("expected cooling on")
		}

		patch := monkey.Patch(time.Since, func(time.Time) time.Duration { return jumpDuration })
		defer patch.Unpatch()

		outputControl.CalculateOutput()

		if l := coolPin.Read(); l != gpio.Low {
			t.Fatal("expected cooling off")
		}
	})
}
