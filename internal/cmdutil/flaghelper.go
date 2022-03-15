package cmdutil

import "github.com/urfave/cli/v2"

func StringFlag(dest *string, name string, usage string) *cli.StringFlag {
	return &cli.StringFlag{
		Name:        name,
		Usage:       usage,
		Value:       *dest,
		Destination: dest,
	}
}

func BoolFlag(dest *bool, name string, usage string) *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        name,
		Usage:       usage,
		Value:       *dest,
		Destination: dest,
	}
}

func IntFlag(dest *int, name string, usage string) *cli.IntFlag {
	return &cli.IntFlag{
		Name:        name,
		Usage:       usage,
		Value:       *dest,
		Destination: dest,
	}
}
