package cmdutil

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/urfave/cli/v2"
)

var (
	instanceName string
	randomName   string
	randomInit   sync.Once
)

func randomInitFunc() {
	var buf [4]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(fmt.Sprintf("cmdutil: unable to pre-compute a random value for the instance name, cause %v", err))
	}
	host, err := os.Hostname()
	if err != nil {
		host = "localhost"
	}
	randomName = fmt.Sprintf("%v-%v-%v", host, os.Getpid(), hex.EncodeToString(buf[:]))
}

func init() {
	randomInit.Do(randomInitFunc)
}

// GetInstanceName uses the provided instance name or returns a
// randomly generated one.
//
// Calling it multiple times will return the same value
func GetInstanceName() string {
	if instanceName == "" {
		return randomName
	}
	return instanceName
}

func InstanceNameFlag() cli.Flag {
	return &cli.StringFlag{
		Name:  "instance-name",
		Usage: "Instance name for this process, leave empty for an auto generated one",
		EnvVars: []string{
			"LSD_INSTANCE_NAME",
			"POD_FULL_NAME",
			"POD_NAME",
		},
		DefaultText: "random",
		Value:       instanceName,
		Destination: &instanceName,
	}
}
