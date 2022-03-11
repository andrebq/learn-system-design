package fleet

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

type (
	// Manager controls a series of sub-processes that compose a fleet
	Manager struct {
		childrenGroup sync.WaitGroup

		binary   string
		baseHost string
		basePort int
	}
)

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
	err := m.startController(ctx)
	if err != nil {
		return err
	}
	m.startStressor(ctx)
	m.startServers(ctx)
}

func (m *Manager) startController(ctx context.Context) error {
	err := m.startCmd(m.childrenGroup.Done, ctx, m.binary, "control-plane", "serve", "--bind", fmt.Sprintf("%v:%v", m.baseHost, m.basePort))
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
		defer done()
		<-ctx.Done()
		cmd.Wait()
	}()
	return nil
}
