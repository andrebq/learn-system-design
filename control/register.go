package control

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Register calls the endpoint registration
func Register(ctx context.Context, controlEndpoint string, name string, service string, publicEndpoint string) error {
	controlEndpoint = strings.TrimRight(controlEndpoint, "/")
	body := Server{
		Name:     name,
		Service:  service,
		Endpoint: publicEndpoint,
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%v/register/server/%v/service/%v", controlEndpoint, name, service), bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("control: unable to register server %v.%v at %v with endpoint %v. Status %v", name, service, controlEndpoint, publicEndpoint, res.Status)
	}
	return nil
}
