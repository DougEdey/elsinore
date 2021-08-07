package devices

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
)

var outpins []*OutPin = nil

// OutPin represents a stored output pin with a friendly name
type OutPin struct {
	gorm.Model

	Identifier   string
	FriendlyName string
	PinIO        gpio.PinIO `gorm:"-"`
	onTime       *time.Time
	offTime      *time.Time
}

func (op *OutPin) off() bool {
	if op == nil {
		return false
	}

	if op.PinIO == nil {
		log.Warn().Msgf("Resetting off %v", op.Identifier)
		err := op.reset()
		if err != nil {
			return false
		}
	}

	if op.PinIO == nil {
		log.Warn().Msgf("Cannot turn off %v", op.Identifier)
		return false
	}

	if op.offTime != nil && op.PinIO.Read() == gpio.Low {
		return false
	}

	if err := op.PinIO.Out(gpio.Low); err != nil {
		log.Fatal().Err(err).Msgf("Failed to set %v to Low (off)", op.FriendlyName)
	}
	curTime := time.Now()
	op.offTime = &curTime
	op.onTime = nil
	return true
}

func (op *OutPin) on() bool {
	if op == nil {
		return false
	}

	if op.PinIO == nil {
		log.Warn().Msgf("Rsetting on %v", op.Identifier)
		err := op.reset()
		if err != nil {
			return false
		}
	}

	if op.PinIO == nil {
		log.Warn().Msgf("Cannot turn on %v", op.Identifier)
		return false
	}

	if op.onTime != nil && op.PinIO.Read() == gpio.High {
		return false
	}

	if err := op.PinIO.Out(gpio.High); err != nil {
		log.Fatal().Err(err).Msgf("Failed to set %v to High (on)", op.FriendlyName)
	}

	if op.PinIO.Read() != gpio.High {
		log.Warn().Msg("Failed to turn pin on! resetting and trying again")
		if err := op.PinIO.Out(gpio.High); err != nil {
			log.Fatal().Err(err).Msgf("Failed to set %v to High (on)", op.FriendlyName)
		}
		if op.PinIO.Read() != gpio.High {
			log.Warn().Msg("Failed to turn pin on!")
		}
	}
	curTime := time.Now()
	op.offTime = nil
	op.onTime = &curTime
	return true
}

func (op *OutPin) reset() error {
	if len(strings.TrimSpace(op.Identifier)) == 0 {
		if op != nil {
			op.off()
		}
		return nil
	}

	if op.PinIO == nil {
		op.PinIO = gpioreg.ByName(op.Identifier)
		if op.PinIO == nil {
			log.Error().Msgf("No Pin for %v!\n", op.Identifier)
			return fmt.Errorf("no pin for %v", op.Identifier)
		}
	}
	log.Warn().Msgf("Reset %v", op.Identifier)
	op.off()
	return nil
}

func (op *OutPin) update(identifier string) {
	if len(strings.TrimSpace(identifier)) == 0 {
		err := op.reset()
		if err != nil {
			log.Warn().Err(err)
		}
	} else if op.Identifier != identifier {
		log.Warn().Msgf("Updating identifier %v", identifier)
		err := op.reset()
		if err != nil {
			log.Warn().Err(err)
		}

		op.PinIO = nil
		op.Identifier = identifier

		err = op.reset()
		if err != nil {
			log.Warn().Err(err)
		}
	}
}

// GpioInUse - Returns true if the GPIO specified is in use already
func GpioInUse(identifier string) bool {
	for _, outpin := range outpins {
		if strings.EqualFold(outpin.Identifier, identifier) {
			return true
		}
	}
	return false
}

func deleteOutpin(outpin *OutPin) {
	if outpin == nil {
		return
	}

	for i, o := range outpins {
		if o == outpin {
			outpins[i] = outpins[len(outpins)-1]
			outpins = outpins[:len(outpins)-1]
			break
		}
	}
}

func createOutpin(identifier string, friendlyName string) (*OutPin, error) {
	if GpioInUse(identifier) {
		return nil, fmt.Errorf("GPIO '%v' is already in use", identifier)
	}
	newPin := &OutPin{Identifier: identifier, FriendlyName: friendlyName}
	outpins = append(outpins, newPin)
	return newPin, nil
}
