//go:build windows

package main

import (
	"atc_freq/internal/app"
	"atc_freq/internal/sim"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

func main() {
	connection, err := sim.NewConnection()
	if err != nil {
		fmt.Printf("Can't create application: %v\n", err)
		os.Exit(1)
	}

	coreApp := app.NewApp(connection)

	cliApp := &cli.App{
		Name:  "cli",
		Usage: "Get airport frequencies and weather information from MSFS",
		Commands: []*cli.Command{
			{
				Name:        "freq",
				Usage:       "Get all frequencies for an airfield",
				ArgsUsage:   "<ICAO>",
				Action:      freq(coreApp),
				Description: "Retrieves and displays all available frequencies for the specified airport.\n\n   Example:\n      atc_freq freq EDDB",
			},
			{
				Name:        "weather",
				Usage:       "Get weather at waypoints",
				ArgsUsage:   "<waypoint1,waypoint2,...>",
				Action:      weather(coreApp),
				Description: "Retrieves weather information for a comma-separated list of waypoints.\n\n   Example:\n      atc_freq weather EDDB,UUMI,KJFK",
			},
			{
				Name:        "clouds",
				Usage:       "Get cloud density around an airport",
				ArgsUsage:   "<ICAO>",
				Action:      clouds(coreApp),
				Description: "Retrieves cloud density information around the specified airport.\n\n   Example:\n      atc_freq clouds EDDB",
			},
		},
	}

	if err := cliApp.Run(os.Args); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func freq(coreApp *app.App) cli.ActionFunc {
	return func(cliContext *cli.Context) error {
		return freqCommand(cliContext, coreApp)
	}
}

func freqCommand(cliContext *cli.Context, app *app.App) error {
	if cliContext.NArg() != 1 {
		return fmt.Errorf("requires exactly one ICAO code argument")
	}

	icao := cliContext.Args().Get(0)

	freqs, err := app.GetFrequencies(icao)
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

func weather(coreApp *app.App) cli.ActionFunc {
	return func(cliContext *cli.Context) error {
		return weatherCommand(cliContext, coreApp)
	}
}

func weatherCommand(ctx *cli.Context, coreApp *app.App) error {
	if ctx.NArg() != 1 {
		return fmt.Errorf("requires exactly one argument (comma-separated waypoints)")
	}

	waypointsStr := ctx.Args().Get(0)
	waypoints := strings.Split(waypointsStr, ",")

	weatherData, err := coreApp.GetWeather(waypoints)
	if err != nil {
		return err
	}

	fmt.Println("Weather information:")
	for wp, weather := range weatherData {
		fmt.Printf("  %s:\n", wp)
		fmt.Printf("    Visibility: %d SM\n", weather.Visibility)
		if len(weather.Clouds) > 0 {
			fmt.Printf("    Clouds:\n")
			for _, cloud := range weather.Clouds {
				fmt.Printf("      %s at %d ft\n", cloud.Coverage, cloud.Base)
			}
		} else {
			fmt.Printf("    Clouds: Clear\n")
		}
		fmt.Printf("    Raw: %s\n", weather.RawMetar)
	}

	return nil
}

func clouds(coreApp *app.App) cli.ActionFunc {
	return func(cliContext *cli.Context) error {
		return cloudsCommand(cliContext, coreApp)
	}
}

func cloudsCommand(ctx *cli.Context, coreApp *app.App) error {
	if ctx.NArg() != 1 {
		return fmt.Errorf("requires exactly one ICAO code argument")
	}

	icao := ctx.Args().Get(0)

	// TODO: Call coreApp method when implemented
	fmt.Printf("Cloud density for %s:\n", strings.ToUpper(icao))
	fmt.Printf("  (stub - not yet implemented)\n")

	return nil
}
