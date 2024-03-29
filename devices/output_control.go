package devices

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var contextDecl, cancelFunc = context.WithCancel(context.Background())

// Context - The global context object
var Context = contextDecl

// CancelFunc - Call this to shutdown the app
var CancelFunc = cancelFunc

// OutputControl is a basic struct to handle heating outputs with a duty cyclke
type OutputControl struct {
	gorm.Model
	HeatOutput *OutPin
	CoolOutput *OutPin
	DutyCycle  int64 `gorm:"-"`
	CycleTime  int64 `gorm:"-"`
}

// RegisterGpios - Register the outpins with the master list
func (o *OutputControl) RegisterGpios() {
	if o == nil {
		return
	}

	if o.HeatOutput != nil {
		outpins = append(outpins, o.HeatOutput)
	}
	if o.CoolOutput != nil {
		outpins = append(outpins, o.CoolOutput)
	}
}

// AfterDelete - After deleting a switch, remove the output pin
func (o *OutputControl) AfterDelete(tx *gorm.DB) {
	deleteOutpin(o.HeatOutput)
	deleteOutpin(o.CoolOutput)
}

// Reset - Reset the output pins
func (o *OutputControl) Reset() {
	if o == nil {
		return
	}
	if o.HeatOutput != nil {
		err := o.HeatOutput.reset()
		if err != nil {
			log.Warn().Err(err)
		}
	}
	if o.CoolOutput != nil {
		err := o.CoolOutput.reset()
		if err != nil {
			log.Warn().Err(err)
		}
	}
}

// UpdateGpios - Update the heating and cooling outputs to their new pins
func (o *OutputControl) UpdateGpios(parentName string, heatGpio string, coolGpio string) error {
	// update the heating pin
	emptyHeatGpio := len(strings.TrimSpace(heatGpio)) == 0
	emptyCoolGpio := len(strings.TrimSpace(coolGpio)) == 0
	if o.HeatOutput == nil && !emptyHeatGpio {
		newPin, err := createOutpin(heatGpio, fmt.Sprintf("%v Heating", parentName))
		if err != nil {
			return err
		}
		o.HeatOutput = newPin
	} else if o.HeatOutput != nil {
		err := o.HeatOutput.update(heatGpio)
		if err != nil {
			return err
		}
		if emptyHeatGpio {
			err := o.HeatOutput.reset()
			if err != nil {
				log.Warn().Err(err)
			}
			o.HeatOutput = nil
		}
	}
	// update the cooling pin
	if o.CoolOutput == nil && !emptyCoolGpio {
		newPin, err := createOutpin(coolGpio, fmt.Sprintf("%v Cooling", parentName))
		if err != nil {
			return err
		}
		o.CoolOutput = newPin
	} else if o.CoolOutput != nil {
		err := o.CoolOutput.update(coolGpio)
		if err != nil {
			return err
		}
		if emptyCoolGpio {
			err := o.CoolOutput.reset()
			if err != nil {
				log.Warn().Err(err)
			}
			o.CoolOutput = nil
		}
	}
	return nil
}

// CalculateOutput - Turn on and off the output pin for this output control depending on the duty cycle
func (o *OutputControl) CalculateOutput() {
	cycleSeconds := math.Abs(float64(o.CycleTime*o.DutyCycle) / 100)
	if cycleSeconds == 0 {
		o.HeatOutput.off()
		o.CoolOutput.off()
	} else if o.DutyCycle == 100 {
		o.CoolOutput.off()
		if o.HeatOutput.on() {
			log.Info().Msgf("Turning on Heat Output (%v) for 100%% duty cycle", o.HeatOutput.FriendlyName)
		}
	} else if o.DutyCycle == -100 {
		o.HeatOutput.off()
		if o.CoolOutput.on() {
			log.Info().Msgf("Turning on Cool Output (%v) for -100%% duty cycle", o.CoolOutput.FriendlyName)
		}
	} else if o.DutyCycle > 0 {
		o.CoolOutput.off()
		if o.HeatOutput.onTime != nil {
			// it's on, do we need to turn it off?
			changedAt := time.Since(*o.HeatOutput.onTime)
			if changedAt.Seconds() > float64(cycleSeconds) {
				log.Info().Msgf("Heat output (%v) turning off after %v seconds", o.HeatOutput.FriendlyName, changedAt.Seconds())
				o.HeatOutput.off()
			}
		} else if o.HeatOutput.offTime != nil {
			// it's off, do we need to turn it on?
			changedAt := time.Since(*o.HeatOutput.offTime)
			offSeconds := float64(o.CycleTime) - cycleSeconds
			if changedAt.Seconds() >= offSeconds {
				log.Info().Msgf("Heat output (%v) turning on after %v seconds", o.HeatOutput.FriendlyName, changedAt.Seconds())
				o.HeatOutput.on()
			}
		} else {
			log.Info().Msgf("Heat output has no on or off time! %v", o.HeatOutput.Identifier)
			o.HeatOutput.off()
		}
	} else if o.DutyCycle < 0 {
		o.HeatOutput.off()

		if o.CoolOutput.onTime != nil {
			// it's on, do we need to turn it off?
			changedAt := time.Since(*o.CoolOutput.onTime)
			if changedAt.Seconds() > float64(cycleSeconds) {
				log.Info().Msgf("Cool output (%v) turning off after %v seconds\n", o.CoolOutput.FriendlyName, changedAt.Seconds())
				o.CoolOutput.off()
			}
		} else if o.CoolOutput.offTime != nil {
			// it's off, do we need to turn it on?
			changedAt := time.Since(*o.CoolOutput.offTime)
			offSeconds := float64(o.CycleTime) - cycleSeconds
			if changedAt.Seconds() >= offSeconds {
				log.Info().Msgf("Cool output (%v) turning on after %v seconds\n", o.CoolOutput.FriendlyName, changedAt.Seconds())
				o.CoolOutput.on()
			}
		}
	}
}

// RunControl -> Run the output controller for a heating output
func (o *OutputControl) RunControl(quit chan struct{}) {
	log.Info().Msgf("Starting output control")
	o.Reset()
	duration, err := time.ParseDuration("10ms")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse 10ms as a duration")
	}

	ticker := time.NewTicker(duration)

	for {
		select {
		case <-ticker.C:
			o.CalculateOutput()
		case <-quit:
			o.Reset()
			ticker.Stop()
			log.Info().Msg("Stop")
			return
		case <-Context.Done():
			o.Reset()
			log.Info().Msg("Done")
			return
		}
	}
}
