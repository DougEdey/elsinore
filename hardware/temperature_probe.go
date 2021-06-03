package hardware

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"periph.io/x/periph/conn/onewire"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/devices/ds18b20"
	"periph.io/x/periph/experimental/host/netlink"
)

var probes = make(map[string]*TemperatureProbe)

// TemperatureProbe holds data that represents a physical temperature probe
// PhysAddr -> The Hex address of the probe on the filesystem
// Address -> The unsigned int value for the readings
// Reading -> The actual reading as a Physic.Temperature
type TemperatureProbe struct {
	PhysAddr   string
	Address    onewire.Address
	ReadingRaw physic.Temperature
	Updated    time.Time
}

// UpdateTemperature Set the temperature on the Temperature Probe from a string
func (t *TemperatureProbe) UpdateTemperature(newTemp string) error {
	return t.ReadingRaw.Set(newTemp)
}

// Reading The current temperature reading for the probe
func (t *TemperatureProbe) Reading() string {
	if t == nil {
		return ""
	}
	return t.ReadingRaw.String()
}

// GetTemperature -> Get the probe object for a physical address
func GetTemperature(physAddr string) *TemperatureProbe {
	return probes[physAddr]
}

// GetProbes -> Get all the probes
func GetProbes() []*TemperatureProbe {
	values := make([]*TemperatureProbe, len(probes))
	i := 0
	for _, v := range probes {
		values[i] = v
		i++
	}
	return values
}

// ReadAddresses -> Update the TemperatureProbes with the current value from the device
func ReadAddresses(oneBus *netlink.OneWire, messages *chan string) {
	for _, probe := range probes {
		defer func() {
			if err := recover(); err != nil {
				log.Error().Msgf("Error reading temperature for %v", probe.PhysAddr)
			}
		}()
		// init ds18b20
		sensor, _ := ds18b20.New(oneBus, probe.Address, 10)
		err := ds18b20.ConvertAll(oneBus, 10)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to update probe %v", probe.PhysAddr)
			continue
		}

		temp, err := sensor.LastTemp()
		if err != nil {
			log.Error().Err(err).Msgf("Failed to get the last temp for %v", probe.PhysAddr)
			continue
		}

		probe := probes[probe.PhysAddr]
		probe.Updated = time.Now()
		probe.ReadingRaw = temp
		if messages != nil {
			*messages <- fmt.Sprintf("Reading device %v: %v", probe.PhysAddr, temp)
		}
	}
}

// ReadTemperatures Read the temperatures on an infinite ticker loop
func ReadTemperatures(m *chan string, quit chan struct{}) {
	if m != nil {
		defer close(*m)
	}
	log.Info().Msgf("Reading temps.")

	oneBus, err := netlink.New(001)
	if err != nil {
		log.Printf("Could not open Netlink host: %v", err)
	}
	defer oneBus.Close()

	// get 1wire address
	addresses, _ := oneBus.Search(false)
	log.Info().Msgf("Reading temps from %v devices.", len(addresses))
	for _, address := range addresses {
		a := strconv.FormatUint(uint64(address), 16)

		for len(a) < 16 {
			// O(n²) but since digits is expected to run for a few loops, it doesn't
			// matter.
			a = "0" + a
		}
		addrBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(addrBytes, uint64(address))
		physAddr := "" + hex.EncodeToString(addrBytes[0:1]) + "-" + hex.EncodeToString(reverse(addrBytes[1:7]))
		log.Info().Msgf("Found %v", physAddr)
		probes[physAddr] = &TemperatureProbe{
			PhysAddr: physAddr,
			Address:  address,
		}
	}
	duration, err := time.ParseDuration("5s")
	if err != nil {
		log.Fatal().Err(err)
	}
	ticker := time.NewTicker(duration)

	for {
		select {
		case <-ticker.C:
			ReadAddresses(oneBus, m)
		case <-quit:
			ticker.Stop()
			log.Info().Msg("Stop")
			return
		}
	}
}

// LogTemperatures -> Write the Temperatures to standard output
func LogTemperatures(messages chan string) {
	for {
		for message := range messages {
			log.Info().Msg(message)
		}
	}
}

// SetProbe -> Used to set a probe in the master list
func SetProbe(probe *TemperatureProbe) {
	probes[probe.PhysAddr] = probe
}

func reverse(arr []byte) []byte {
	for i := 0; i < len(arr)/2; i++ {
		j := len(arr) - i - 1
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}
