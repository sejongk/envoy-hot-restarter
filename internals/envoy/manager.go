package envoy

import (
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	TERM_WAIT_SECONDS int = 30
)

type EnvoyManager struct {
	Envoys          []*Envoy
	WaitGroup       sync.WaitGroup
	StoppedEnvoy    chan *Envoy
	StartedEnvoy    chan *Envoy
	RestartEpoch    int
	EnvoyExecPath   string
	EnvoyConfigpath string
	EnvoyNodeId     string
	Done            bool
}

func (manager *EnvoyManager) Start() {
	logger.Info("starting envoy manager")
	manager.Done = false

	go manager.Process()
}

func (manager *EnvoyManager) Process() {
	for {
		select {
		case envoy := <-manager.StartedEnvoy:
			manager.Envoys = append(manager.Envoys, envoy)
		case envoy := <-manager.StoppedEnvoy:
			for i, e := range manager.Envoys {
				if e == envoy {
					manager.Envoys = append(manager.Envoys[:i], manager.Envoys[i+1:]...)
				}
			}
			if envoy.Error != nil {
				manager.Stop()
			}
		}
	}
}

func (manager *EnvoyManager) Stop() {
	if !manager.Done {
		manager.Done = true
		logger.Info("terminating envoy manager and all envoys")
		go manager.TermAllEnvoys()
	}
}

func (manager *EnvoyManager) StartNewEnvoy() {
	manager.RestartEpoch += 1
	logger.Info("starting new envoy at epoch ", manager.RestartEpoch)
	manager.WaitGroup.Add(1)

	envoyCmd := exec.Command(manager.EnvoyExecPath, "-c", manager.EnvoyConfigpath, "--restart-epoch", strconv.Itoa(manager.RestartEpoch), "--service-node", manager.EnvoyNodeId)
	envoyCmd.Stdout = os.Stdout
	envoyCmd.Stderr = os.Stderr

	envoy := &Envoy{
		Manager:      manager,
		Cmd:          envoyCmd,
		RestartEpoch: manager.RestartEpoch,
		Error:        nil,
	}
	envoy.Manager.StartedEnvoy <- envoy
	envoy.Start()
}

func (manager *EnvoyManager) TermAllEnvoys() {
	for _, envoy := range manager.Envoys {
		logger.Info("gracefully terminating envoy pid ", envoy.Cmd.Process.Pid)
		err := envoy.Cmd.Process.Signal(syscall.SIGTERM)
		if err != nil {
			logger.Error("failed to terminate gracefully envoy pid ", envoy.Cmd.Process.Pid)
		}
	}

	select {
	case <-time.After(time.Second * time.Duration(TERM_WAIT_SECONDS)):
		logger.Info("graceful termination timeout, force kill remain envoys")
		for _, envoy := range manager.Envoys {
			logger.Info("forcefully killing envoy pid ", envoy.Cmd.Process.Pid)
			err := envoy.Cmd.Process.Kill()
			if err != nil {
				logger.Error("failed to kill forcefully envoy pid ", envoy.Cmd.Process.Pid)
			}
		}
	}
}

func (manager *EnvoyManager) Wait() {
	manager.WaitGroup.Wait()
}
