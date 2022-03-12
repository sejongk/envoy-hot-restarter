package envoy

import (
	"os/exec"
)

type Envoy struct {
	Manager      *EnvoyManager
	Cmd          *exec.Cmd
	RestartEpoch int
	Error        error
}

func (envoy *Envoy) Start() {
	go func() {
		defer envoy.Manager.WaitGroup.Done()
		envoy.Error = envoy.Cmd.Start()

		if envoy.Error != nil {
			logger.Error(envoy.Error, "failed to start new envoy")
			envoy.Manager.StoppedEnvoy <- envoy
		}

		envoy.Error = envoy.Cmd.Wait()
		envoy.Manager.StoppedEnvoy <- envoy
		if envoy.Error != nil {
			logger.Error("abnormally terminated envoy pid ", envoy.Cmd.Process.Pid)
		} else {
			logger.Info("successfully terminated envoy pid ", envoy.Cmd.Process.Pid)
		}

	}()
}
