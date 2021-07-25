package devices

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/dougedey/elsinore/database"
	"github.com/dougedey/elsinore/graph/model"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"periph.io/x/periph/conn/gpio"
)

var switches []*Switch = nil

// AllSwitches returns all the switches, loading from the Database if none are configured
func AllSwitches() []*Switch {
	if switches == nil && database.FetchDatabase() != nil {
		log.Info().Msg("Switches array is nil, checking the database...")
		database.FetchDatabase().Debug().Preload(clause.Associations).Find(&switches)
	}
	return switches
}

// ShutdownAllSwitches - Turn off all switches that are configured, does not cover output control pins
func ShutdownAllSwitches() {
	if len(switches) == 0 {
		log.Info().Msg("No switches to shutdown.\n")
		return
	}

	log.Info().Msgf("Shutting down %v switches...\n", len(outpins))
	for _, s := range switches {
		log.Info().Msgf("Shutting down %v...", s.Output.FriendlyName)
		s.Off()
		log.Info().Msgf("Done %v!\n", s.Output.FriendlyName)
	}
}

// FindSwitchByID - Find a temperature controller by id, preloading everything
func FindSwitchByID(id string) *Switch {
	intID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil
	}

	for i := range switches {
		log.Info().Msgf("Comparing %v to %v\n", switches[i].ID, uint(intID))
		if switches[i].ID == uint(intID) {
			return switches[i]
		}
	}
	var existingSwitch *Switch = nil
	if database.FetchDatabase() == nil {
		log.Printf("No Database configured, cannot lookup by Id")
		return nil
	}
	result := database.FetchDatabase().Debug().Preload(clause.Associations).First(&existingSwitch, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}

	return existingSwitch
}

// AfterDelete - After deleting a switch, remove the output pin
func (s *Switch) AfterDelete(tx *gorm.DB) {
	deleteOutpin(s.Output)
}

// DeleteSwitchByID - Does what it says on the tin, clearing up the Outpin too
func DeleteSwitchByID(id string) (*Switch, error) {
	existingSwitch := FindSwitchByID(id)
	if existingSwitch == nil {
		return nil, fmt.Errorf("no switch found with id '%v'", id)
	}
	database.FetchDatabase().Debug().Delete(&existingSwitch)
	for i, s := range switches {
		if s == existingSwitch {
			switches[i] = switches[len(switches)-1]
			switches = switches[:len(switches)-1]
			break
		}
	}
	return existingSwitch, nil
}

// CreateSwitch - Create a new switch, checking for the GPIO/Name already existing
func CreateSwitch(identifier string, friendlyName string) (*Switch, error) {
	for _, s := range switches {
		if strings.EqualFold(s.Name(), friendlyName) {
			return nil, fmt.Errorf("switch '%v' already exists", friendlyName)
		}
	}

	newPin, err := createOutpin(identifier, friendlyName)
	if err != nil {
		return nil, err
	}

	newSwitch := Switch{Output: newPin}
	switches = append(switches, &newSwitch)
	database.FetchDatabase().Debug().Save(&newSwitch)
	return &newSwitch, nil
}

// Switch - Represents a thin wrapper around an OutPin to define a switch
type Switch struct {
	gorm.Model
	OutputID uint
	Output   *OutPin `gorm:"ForeignKey:OutputID"`
	Inverted bool
}

// Reset - Turn off the switch if configured
func (s *Switch) Reset() {
	if s == nil || s.Output == nil {
		return
	}

	s.Off()
}

// On - Switch on the output pin, if it's inverted, the pin goes to off
func (s *Switch) On() {
	if s.Output == nil {
		return
	}

	if s.Inverted {
		s.Output.off()
	} else {
		s.Output.on()
	}
}

// Off - Switch off the output pin, if it's inverted, the pin goes to on
func (s *Switch) Off() {
	if s.Output == nil {
		return
	}

	if s.Inverted {
		s.Output.on()
	} else {
		s.Output.off()
	}
}

// Gpio - Get the GPIO
func (s *Switch) Gpio() string {
	return s.Output.Identifier
}

// Name - Get the Name
func (s *Switch) Name() string {
	return s.Output.FriendlyName
}

// State - Returns on if this switch is on
func (s *Switch) State() model.SwitchMode {
	if s.Output.PinIO.Read() == s.onState() {
		return model.SwitchModeOn
	}
	return model.SwitchModeOff
}

func (s *Switch) onState() gpio.Level {
	if s.Inverted {
		return gpio.Low
	}
	return gpio.High
}
