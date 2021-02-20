package elsinore

import (
	"testing"

	"periph.io/x/periph/conn/onewire"
)

func TestGetTemperature(t *testing.T) {
	emptyResult := GetTemperature("NotAnAddress")

	if emptyResult != nil {
		t.Errorf("GetTemperature with an invalid address string failed, expected nil, got %v", emptyResult)
	}

	realAddress := "ARealAddress"
	probes[realAddress] = &TemperatureProbe{
		PhysAddr: realAddress,
		Address: onewire.Address(12345),
	}

	actualProbe := GetTemperature(realAddress)

	if actualProbe == nil {
		t.Errorf("GetTemperature returned a nil object for %v", realAddress)
	}
}