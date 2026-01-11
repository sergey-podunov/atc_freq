//go:build windows

package main

import (
	"atc_freq/internal/sim"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "atc_freq",
		Usage: "Get airport frequencies and weather information from MSFS",
		Commands: []*cli.Command{
			{
				Name:        "freq",
				Usage:       "Get all frequencies for an airfield",
				ArgsUsage:   "<ICAO>",
				Action:      freqCommand,
				Description: "Retrieves and displays all available frequencies for the specified airport.\n\n   Example:\n      atc_freq freq EDDB",
			},
			{
				Name:        "weather",
				Usage:       "Get weather at waypoints",
				ArgsUsage:   "<waypoint1,waypoint2,...>",
				Action:      weatherCommand,
				Description: "Retrieves weather information for a comma-separated list of waypoints.\n\n   Example:\n      atc_freq weather EDDB,UUMI,KJFK",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func freqCommand(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return fmt.Errorf("requires exactly one ICAO code argument")
	}

	icao := ctx.Args().Get(0)

	client, err := sim.NewClient()
	if err != nil {
		return fmt.Errorf("failed to load Connection: %w", err)
	}

	service := sim.NewService(client)

	freqs, err := service.GetFrequency(icao)
	if err != nil {
		return err
	}

	if len(freqs) == 0 {
		fmt.Printf("No frequencies found for %s\n", strings.ToUpper(icao))
		return nil
	}

	fmt.Printf("Frequencies for %s:\n", strings.ToUpper(icao))
	for _, f := range freqs {
		fmt.Printf("  %-10s %8.3f MHz  %s\n", f.Type, f.MHz, f.Name)
	}

	return nil
}

func weatherCommand(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return fmt.Errorf("requires exactly one argument (comma-separated waypoints)")
	}

	waypointsStr := ctx.Args().Get(0)
	waypoints := strings.Split(waypointsStr, ",")

	client, err := sim.NewClient()
	if err != nil {
		return fmt.Errorf("failed to load Connection: %w", err)
	}

	service := sim.NewService(client)

	weatherData, err := service.GetWeather(waypoints)
	if err != nil {
		return err
	}

	fmt.Println("Weather information:")
	for wp, data := range weatherData {
		fmt.Printf("  %s: %s\n", wp, data)
	}

	return nil
}
