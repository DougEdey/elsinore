// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

// The new settings for hysteria mode
type HysteriaSettingsInput struct {
	// Indicates if these settings have been configured yet
	Configured *bool `json:"configured"`
	// When this temperature is hit, turn on the cooling output
	MaxTemp *string `json:"maxTemp"`
	// When this temperature is hit, turn on the heating output
	MinTemp *string `json:"minTemp"`
	// The minimum amount of time to turn the outputs on for.
	MinTime *int `json:"minTime"`
}

// The new manual settings for this controller
type ManualSettingsInput struct {
	// Indicates if these settings have been configured yet
	Configured *bool `json:"configured"`
	// The time for one duty cycle in seconds
	CycleTime *int `json:"cycleTime"`
	// The manual duty cycle percentage for this controller
	DutyCycle *int `json:"dutyCycle"`
}

// The settings for heating or cooling on a temperature controller
type PidSettingsInput struct {
	// Indicates if these settings have been configured yet
	Configured *bool `json:"configured"`
	// The automatic cycle time in seconds
	CycleTime *int `json:"cycleTime"`
	// The minimum delay between turning an output on and off in seconds
	Delay *int `json:"delay"`
	// The derivative calculation value
	Derivative *float64 `json:"derivative"`
	// The integral calculation value
	Integral *float64 `json:"integral"`
	// The proportional calculation value
	Proportional *float64 `json:"proportional"`
	// The friendly name of the GPIO Value
	Gpio *string `json:"gpio"`
}

// The new settings for this brewery
type SettingsInput struct {
	// The new brewery name (blank for no change)
	BreweryName *string `json:"breweryName"`
}

type SwitchSettingsInput struct {
	// The Id of the switch, if no ID, create a new switch
	ID *string `json:"id"`
	// The new Name for the switch (required during switch creation)
	Name *string `json:"name"`
	// The new GPIO for the switch (required during switch creation)
	Gpio *string `json:"gpio"`
	// The new state for the switch
	State *SwitchMode `json:"state"`
}

// A device that reads a temperature and is assigned to a temperature controller
type TempProbeDetails struct {
	// The ID of an object
	ID string `json:"id"`
	// The physical address of this probe
	PhysAddr *string `json:"physAddr"`
	// The value of the reading
	Reading *string `json:"reading"`
	// The friendly name of this probe
	Name *string `json:"name"`
	// The time that this reading was updated
	Updated *time.Time `json:"updated"`
}

// Used to configure a controller
type TemperatureControllerSettingsInput struct {
	// The controller Id
	ID string `json:"id"`
	// The name of the controller.
	Name *string `json:"name"`
	// The new mode for the controller
	Mode *ControllerMode `json:"mode"`
	// The PID Settings for the cooling output
	CoolSettings *PidSettingsInput `json:"coolSettings"`
	// The PID settings for the heating output
	HeatSettings *PidSettingsInput `json:"heatSettings"`
	// The hysteria settings for controlling this temperature controller
	HysteriaSettings *HysteriaSettingsInput `json:"hysteriaSettings"`
	// The manual settings for this temperature controller
	ManualSettings *ManualSettingsInput `json:"manualSettings"`
	// The target for auto mode
	SetPoint *string `json:"setPoint"`
}

// A device that reads a temperature
type TemperatureProbe struct {
	// The physical address of this probe
	PhysAddr *string `json:"physAddr"`
	// The value of the reading
	Reading *string `json:"reading"`
	// The time that this reading was updated
	Updated *time.Time `json:"updated"`
}

type SwitchMode string

const (
	SwitchModeOn  SwitchMode = "on"
	SwitchModeOff SwitchMode = "off"
)

var AllSwitchMode = []SwitchMode{
	SwitchModeOn,
	SwitchModeOff,
}

func (e SwitchMode) IsValid() bool {
	switch e {
	case SwitchModeOn, SwitchModeOff:
		return true
	}
	return false
}

func (e SwitchMode) String() string {
	return string(e)
}

func (e *SwitchMode) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = SwitchMode(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid SwitchMode", str)
	}
	return nil
}

func (e SwitchMode) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
