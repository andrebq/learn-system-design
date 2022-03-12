package stress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/andrebq/learn-system-design/internal/cmdutil"
	"github.com/andrebq/learn-system-design/stress"
	"github.com/urfave/cli/v2"
)

func Cmd() *cli.Command {
	return &cli.Command{
		Name:  "stress",
		Usage: "Contains sub-commands to interact with the stress test server",
		Subcommands: []*cli.Command{
			serveCmd(),
			startCmd(),
		},
	}
}

func serveCmd() *cli.Command {
	var bind string = "127.0.0.1:9001"
	var publicEndpoint string
	var controlEndpoint string = "http://127.0.0.1:9000"
	return &cli.Command{
		Name:  "serve",
		Usage: "Serve the API that allows clients to run stress tests",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "bind",
				Usage:       "Address to bind and wait for stress commands",
				EnvVars:     []string{"LSD_STRESS_SERVE_BIND"},
				Destination: &bind,
				Value:       bind,
			},
			&cli.StringFlag{
				Name: "public-endpoint",
				Usage: `Endpoint used when sending registration information to control plane.

Then endpoint of the control plane itself is defined by the 'control-endpoint' argument.

This argument is mandatory!
`,
				EnvVars:     []string{"LSD_STRESSOR_SERVE_PUBLIC_ENDPOINT"},
				Value:       publicEndpoint,
				Destination: &publicEndpoint,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "control-endpoint",
				Usage:       "Base endpoint used to register this instance in the control plane",
				EnvVars:     []string{"LSD_STRESSOR_SERVE_CONTROL_ENDPOINT"},
				Value:       controlEndpoint,
				Destination: &controlEndpoint,
			},
		},
		Action: func(ctx *cli.Context) error {
			h := stress.Handler(ctx.Context, cmdutil.GetInstanceName(), controlEndpoint, publicEndpoint)
			return cmdutil.RunHTTPServer(ctx.Context, h, bind)
		},
	}
}

func startCmd() *cli.Command {
	var target string
	var method string
	var duration time.Duration = time.Second * 5
	var rps int = 1000
	var workers int = 10
	var stressorEndpoint string = "http://127.0.0.1:9001/start-test"
	return &cli.Command{
		Name:  "start",
		Usage: "Interacts with the stress test API to start/stop tests",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "stressor",
				Usage:       "URL where the stress test server is running (this is the one that will perform the test)",
				Destination: &stressorEndpoint,
				Value:       stressorEndpoint,
			},
			&cli.StringFlag{
				Name:        "target",
				Usage:       "Target URL to stress",
				Destination: &target,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "method",
				Usage:       "Method to use, defaults to GET",
				Destination: &method,
				Value:       method,
			},
			&cli.DurationFlag{
				Name:        "duration",
				Usage:       "How long should the test last",
				Destination: &duration,
				Value:       duration,
			},
			&cli.IntFlag{
				Name:        "requests-per-second",
				Aliases:     []string{"rps", "rate"},
				Usage:       "How many requests per second should be made (constant pace)",
				Destination: &rps,
				Value:       rps,
			},
			&cli.IntFlag{
				Name:        "workers",
				Usage:       "How many workers to start (more might be created to sustain the desired rate)",
				Destination: &workers,
				Value:       workers,
			},
		},
		Action: func(ctx *cli.Context) error {
			req := stress.StressTest{
				Target:            target,
				Method:            method,
				Workers:           workers,
				Sustain:           duration,
				RequestsPerSecond: rps,
			}
			data, err := json.Marshal(req)
			if err != nil {
				return err
			}
			res, err := http.Post(stressorEndpoint, "application/json", bytes.NewBuffer(data))
			if err != nil {
				return err
			}
			res.Body.Close()
			if res.StatusCode != http.StatusCreated {
				return fmt.Errorf("unexpected status code, got %v", res.Status)
			}
			return nil
		},
	}
}
