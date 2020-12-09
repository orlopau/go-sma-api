package main

import (
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := &cli.App{
		Usage:     "communication tool for renewable energy devices",
		UsageText: "energy-cli <command> [options] [arguments...]",
		Commands: []*cli.Command{
			{
				Name:      "fetch",
				Usage:     "fetches data from SunSpec devices",
				UsageText: "energy-cli fetch --slaveId [<id>] [--addr [<addr>] --addr [<addr>] ...]",
				Description: "Fetches data from SunSpec devices. Supported device types are:\n" +
					" * SunSpec PV Inverters\n" +
					" * SunSpec Battery Inverters\n\n" +
					" If no addresses are specified, the tool will attempt to discover devices in the local network.",
				Aliases: []string{"f"},
				Flags: []cli.Flag{
					&cli.UintFlag{
						Name:     "slaveId",
						Aliases:  []string{"id"},
						Usage:    "slave id to use for modbus connection",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:     "addrs",
						Aliases:  []string{"a"},
						Usage:    "addresses of devices",
						Required: false,
					},
				},
				Action: func(context *cli.Context) error {
					slaveId := context.Uint("slaveId")
					if slaveId > uint(^byte(0)) {
						return cli.Exit("slave id must be in range 0 to 254", 1)
					}

					addrs := context.StringSlice("addrs")
					if !context.IsSet("addrs") || len(addrs) == 0 {
						return toExitCode(discoveryFetch(byte(slaveId)))
					}

					return toExitCode(addressFetch(byte(slaveId), addrs))
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

func toExitCode(err error) cli.ExitCoder {
	if err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}

func discoveryFetch(slaveId byte) error {
	return cli.Exit("Automatic discovery devices is not implemented yet!", 1)
}