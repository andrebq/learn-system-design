package managerctl

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/urfave/cli/v2"
)

func Cmd() *cli.Command {
	var endpoint string = "http://localhost:9001/"
	return &cli.Command{
		Name:  "manager",
		Usage: "Interact with the manager",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "endpoint",
				Usage:       "Base endpoint where the manager is located",
				Aliases:     []string{"manager"},
				Destination: &endpoint,
				Value:       endpoint,
			},
		},
		Subcommands: []*cli.Command{
			uploadCode(&endpoint),
		},
	}
}

func uploadCode(endpoint *string) *cli.Command {
	var serviceType string
	var writeHeaders bool
	return &cli.Command{
		Name:    "upload-code",
		Aliases: []string{"upload"},
		Usage:   "Upload code read from stdin to the selected endpoint",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "service-type",
				Usage:       "Type of the service that should be changed",
				Required:    true,
				Destination: &serviceType,
			},
			&cli.BoolFlag{
				Name:        "write-headers",
				Usage:       "Indicates the HTTP response headres should be written",
				Destination: &writeHeaders,
			},
		},
		Action: func(ctx *cli.Context) error {
			u, err := url.Parse(*endpoint)
			if err != nil {
				return err
			}
			u.Path = path.Join(u.Path, fmt.Sprintf("/code/%v.lua", serviceType))
			req, err := http.NewRequest("PUT", u.String(), os.Stdin)
			if err != nil {
				return err
			}
			req = req.WithContext(ctx.Context)
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if writeHeaders {
				err = res.Header.Write(os.Stdout)
				if err != nil {
					return err
				}
			}
			_, err = os.Stdout.WriteString("\n")
			if err != nil {
				return err
			}
			_, err = io.Copy(os.Stdout, res.Body)
			return err
		},
	}
}
