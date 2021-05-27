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

	testPin := gpiotest.Pin{N: "GPIO21", Num: 10, Fn: "I2C1_SDA"}
	out21 := devices.OutPin{Identifier: "GPIO21", FriendlyName: "GPIO21", PinIO: &testPin}
	outputControl := devices.OutputControl{HeatOutput: &out21, DutyCycle: 50, CycleTime: 4}
	outputControl.Reset()

	t.Run("OutputControl turns off imediately when there's no prior state", func(t *testing.T) {
		if l := testPin.Read(); l != gpio.Low {
			t.Fatal("expected off")
		}

		outputControl.CalculateOutput()

		if l := testPin.Read(); l != gpio.Low {
			t.Fatal("expected off")
		}
	})

	t.Run("OutputControl waits for the off cycle time before turning on for 2s with a 50% duty cycle on a time of 4s", func(t *testing.T) {
		if l := testPin.Read(); l != gpio.Low {
			t.Fatal("expected off")
		}

		patch := monkey.Patch(time.Since, func(time.Time) time.Duration { return jumpDuration })
		defer patch.Unpatch()

		outputControl.CalculateOutput()

		if l := testPin.Read(); l != gpio.High {
			t.Fatal("expected on")
		}
	})

	t.Run("OutputControl waits for the on cycle time before turning off for 2s with a 50% duty cycle on a time of 4s", func(t *testing.T) {
		if l := testPin.Read(); l != gpio.High {
			t.Fatal("expected on")
		}

		patch := monkey.Patch(time.Since, func(time.Time) time.Duration { return jumpDuration })
		defer patch.Unpatch()

		outputControl.CalculateOutput()

		if l := testPin.Read(); l != gpio.Low {
			t.Fatal("expected off")
		}
	})
}
