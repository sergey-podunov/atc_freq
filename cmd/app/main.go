//go:build windows

package main

import (
	"atc_freq/internal/sim"
	"fmt"
	"time"
)

func main() {
	client, err := sim.NewSimClient()
	if err != nil {
		fmt.Println("Failed to load SimConnect:", err)
		return
	}

	// Now you can reuse 'client' for multiple calls
	freqs, err := client.GetAirportFrequencies("EDDB", 10*time.Second)
	if err != nil {
		fmt.Println("ERR:", err)
		return
	}
	for _, f := range freqs {
		fmt.Printf("%-10s %8.3f  %s\n", f.Type, f.MHz, f.Name)
	}

	fmt.Println()
	
	freqs, err = client.GetAirportFrequencies("UUMI", 10*time.Second)
	if err != nil {
		fmt.Println("ERR:", err)
		return
	}
	for _, f := range freqs {
		fmt.Printf("%-10s %8.3f  %s\n", f.Type, f.MHz, f.Name)
	}
}
