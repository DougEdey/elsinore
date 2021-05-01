package model

// ControllerMode Auto, Manual, Off, Hysteria
type ControllerMode string

// The deleted controller
type DeleteTemperatureControllerReturnType struct {
	// The ID of the deleted Controller
	ID string `json:"id"`
	// Temperatures Probes that were associated with this controller
	TemperatureProbes []*string `json:"temperatureProbes"`
}