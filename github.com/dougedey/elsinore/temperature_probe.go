package elsinore

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"time"

	"periph.io/x/periph/conn/onewire"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/devices/ds18b20"
	"periph.io/x/periph/experimental/host/netlink"
	"periph.io/x/periph/host"
)

var probes = make(map[string]*TemperatureProbe)

// TemperatureProbe holds data that represents a physical temperature probe
// PhysAddr -> The Hex address of the probe on the filesystem
// Address -> The unsigned int value for the readings
// Reading -> The actual reading as a Physic.Temperature
type TemperatureProbe struct {
	PhysAddr string             `json:"physAddr"`
	Address  onewire.Address    `json:"address"`
	Reading  physic.Temperature `json:"reading"`
}

// GetTemperature -> Get the probe object for a physical address
func GetTemperature(physAddr string) *TemperatureProbe {
	probe := probes[physAddr]
	log.Printf("Found probe for %v: %v\n", physAddr, probe)
	return probe
}

// ReadAddresses -> Update the TemperatureProbes with the current value from the device
func ReadAddresses(oneBus *netlink.OneWire, messages chan string) {
	for _, probe := range probes {
		// init ds18b20
		sensor, _ := ds18b20.New(oneBus, probe.Address, 10)
		err := ds18b20.ConvertAll(oneBus, 10)
		if err != nil {
			log.Printf("Failed to update probe %v: %v", probe.PhysAddr, err)
		}

		temp, _ := sensor.LastTemp()

		probe := probes[probe.PhysAddr]
		probe.Reading = temp
		messages <- fmt.Sprintf("Reading device %v: %v", probe.PhysAddr, temp)
	}
}

// ReadTemperatures Read the temperatures on an infinite ticker loop
func ReadTemperatures(m chan string) {
	defer close(m)
	fmt.Println("Reading temps.")
	_, err := host.Init()
	if err != nil {
		log.Fatalf("failed to initialize periph: %v", err)
	}

	oneBus, err := netlink.New(001)
	if err != nil {
		log.Printf("Could not open Netlink host: %v", err)
	}
	defer oneBus.Close()

	// get 1wire address
	addresses, _ := oneBus.Search(false)
	fmt.Printf("Reading temps from %v devices.\n", addresses)
	for _, address := range addresses {
		a := strconv.FormatUint(uint64(address), 16)

		for len(a) < 16 {
			// O(n²) but since digits is expected to run for a few loops, it doesn't
			// matter.
			a = "0" + a
		}
		addrBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(addrBytes, uint64(address))
		fmt.Printf("%v", addrBytes)
		physAddr := "" + hex.EncodeToString(addrBytes[0:1]) + "-" + hex.EncodeToString(reverse(addrBytes[1:7]))
		fmt.Printf("Found %v", physAddr)
		probes[physAddr] = &TemperatureProbe{
			PhysAddr: physAddr,
			Address:  address,
		}
	}
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})

	for {
		select {
		case <-ticker.C:
			ReadAddresses(oneBus, m)
		case <-quit:
			ticker.Stop()
			fmt.Println("Stop")
			return
		}
	}
}

// LogTemperatures -> Write the Temperatures to standard output
func LogTemperatures(messages chan string) {
	for {
		for message := range messages {
			fmt.Println(message)
		}
	}
}

func reverse(arr []byte) []byte {
	for i := 0; i < len(arr)/2; i++ {
		j := len(arr) - i - 1
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}
