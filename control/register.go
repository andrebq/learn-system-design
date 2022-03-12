package control

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func RegisterStressor(ctx context.Context, controlEndpoint string, name string, publicEndpoint string) error {
	controlEndpoint = strings.TrimRight(controlEndpoint, "/")
	body := Stressor{
		Name:         name,
		BaseEndpoint: publicEndpoint,
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%v/register/stressor/%v", controlEndpoint, name), bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("control: unable to register stressor %v at %v with endpoint %v. Status %v", name, controlEndpoint, publicEndpoint, res.Status)
	}
	return nil
}

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

// Services returns the list of services that are registered
func Services(ctx context.Context, controlEndpoint string) ([]*Server, error) {
	controlEndpoint = strings.TrimRight(controlEndpoint, "/")
	var registry struct {
		Servers []*Server `json:"servers"`
	}
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%v/", controlEndpoint), nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("control: unable to fetch servers from %v, status %v", controlEndpoint, res.Status)
	}
	err = json.NewDecoder(res.Body).Decode(&registry)
	if err != nil {
		return nil, fmt.Errorf("control: unable to decode response from control, cause %v", err)
	}
	return registry.Servers, nil
}
