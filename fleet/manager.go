package fleet

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/andrebq/learn-system-design/internal/logutil"
	"github.com/rs/zerolog/log"
)

type (
	// Manager controls a series of sub-processes that compose a fleet
	Manager struct {
		childrenGroup sync.WaitGroup

		binary      string
		baseHost    string
		basePort    int
		services    []string
		scriptsBase string
		stressors   int
	}
)

func NewManager(baseBinary, baseHost string, basePort int, scriptsBase string, stressors int, services []string) *Manager {
	return &Manager{
		binary:      baseBinary,
		baseHost:    baseHost,
		basePort:    basePort,
		scriptsBase: scriptsBase,
		stressors:   stressors,
		services:    append([]string(nil), services...),
	}
}

func (m *Manager) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	// lock until the whole fleet has terminated
	defer m.stopFleet(cancel)

	err := m.startFleet(ctx)
	if err != nil {
		return err
	}
	<-ctx.Done()
	return ctx.Err()
}

func (m *Manager) stopFleet(cancel context.CancelFunc) {
	// notify children to stop
	cancel()
	m.childrenGroup.Wait()
}

func (m *Manager) startFleet(ctx context.Context) error {
	controlEndpoint, err := m.startController(ctx)
	if err != nil {
		return err
	}
	for i := 0; i < m.stressors; i++ {
		err = m.startStressor(ctx, controlEndpoint)
		if err != nil {
			return err
		}
	}
	err = m.startServers(ctx, controlEndpoint)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) startServers(ctx context.Context, controlEndpoint string) error {
	for _, service := range m.services {
		err := m.startCmd(m.childrenGroup.Done, ctx, m.binary,
			"--otel.service.name", service,
			"serve", "--bind", fmt.Sprintf("%v:%v", m.baseHost, m.basePort),
			"--handler-file", filepath.Join(m.scriptsBase, service, "handler.lua"),
			"--public-endpoint", fmt.Sprintf("http://%v:%v", m.baseHost, m.basePort),
			"--control-endpoint", controlEndpoint)
		if err != nil {
			return err
		}
		m.childrenGroup.Add(1)
		m.basePort++
	}
	return nil
}

func (m *Manager) startController(ctx context.Context) (string, error) {
	err := m.startCmd(m.childrenGroup.Done, ctx, m.binary, "control-plane", "serve", "--bind", fmt.Sprintf("%v:%v", m.baseHost, m.basePort))
	if err != nil {
		return "", err
	}
	endpoint := fmt.Sprintf("http://%v:%v", m.baseHost, m.basePort)
	m.basePort++
	m.childrenGroup.Add(1)
	return endpoint, nil
}

func (m *Manager) startStressor(ctx context.Context, controlEndpoint string) error {
	err := m.startCmd(m.childrenGroup.Done, ctx, m.binary, "stress", "serve", "--bind", fmt.Sprintf("%v:%v", m.baseHost, m.basePort),
		"--public-endpoint", fmt.Sprintf("http://%v:%v", m.baseHost, m.basePort),
		"--control-endpoint", controlEndpoint)
	if err != nil {
		return err
	}
	m.basePort++
	m.childrenGroup.Add(1)
	return nil
}

func (m *Manager) startCmd(done func(), ctx context.Context, binary string, args ...string) error {
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Stdout = io.Discard
	// lazy approach but it is good enough for our use-case
	cmd.Stderr = os.Stderr
	cmd.Stdin = nil

	err := cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		log := logutil.Acquire(ctx)
		<-ctx.Done()
		log.Info().Str("binary", binary).Strs("args", args).Int("pid", cmd.Process.Pid).Msg("Sending kill signal")
		cmd.Process.Kill()
	}()
	go func() {
		defer done()
		cmd.Wait()
		log.Info().Str("binary", binary).Strs("args", args).Int("pid", cmd.Process.Pid).Int("exitCode", cmd.ProcessState.ExitCode()).Msg("Process done")
	}()
	return nil
}
