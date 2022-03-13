package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/sepaper/envoy-hot-restarter/internals/envoy"
	"github.com/sepaper/envoy-hot-restarter/internals/util"
	"github.com/sepaper/envoy-hot-restarter/internals/watcher"
)

var (
	logger          log.FieldLogger
	envoyConfigPath string
	envoyExecPath   string
	envoyNodeId     string
	sigs            chan os.Signal
	manager         *envoy.EnvoyManager
)

func init() {
	logger = util.GetLogger()
	flag.StringVar(&envoyConfigPath, "envoyConfigPath", "/envoy/envoy-static.yaml", "Envoy config file path")
	flag.StringVar(&envoyExecPath, "envoyExecPath", "/envoy/envoy", "Envoy exec file path")
	flag.StringVar(&envoyNodeId, "envoyNodeId", "node-id", "Envoy service node id")
	flag.Parse()

	sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	manager = &envoy.EnvoyManager{
		RestartEpoch:    -1,
		Envoys:          make([]*envoy.Envoy, 0, 5),
		StoppedEnvoy:    make(chan *envoy.Envoy, 1),
		StartedEnvoy:    make(chan *envoy.Envoy, 1),
		EnvoyExecPath:   envoyExecPath,
		EnvoyConfigpath: envoyConfigPath,
		EnvoyNodeId:     envoyNodeId,
	}
}

type EnvoyConfigFileEventHandler struct {
}

func (h *EnvoyConfigFileEventHandler) On_deleted() error {
	logger.Info("file deleted")
	manager.StartNewEnvoy()
	return nil
}
func (h *EnvoyConfigFileEventHandler) On_created() error {
	return nil
}
func (h *EnvoyConfigFileEventHandler) On_modified() error {
	logger.Info("file modified")
	manager.StartNewEnvoy()
	return nil
}

func startSignalHandler() {
	go func() {
		for {
			select {
			case sig := <-sigs:
				if sig == syscall.SIGINT || sig == syscall.SIGTERM {
					logger.Info("got SIGINT or SIGTERM")
					manager.Stop()
				}
			}
		}
	}()
}

func startWatcher() {
	go func() {
		watcher.Watch(envoyConfigPath, &EnvoyConfigFileEventHandler{})
	}()
}

func main() {
	logger.Info("starting hot-restarter (envoyExecPath:", envoyExecPath, ", envoyConfigPath: ", envoyConfigPath, ")")

	startSignalHandler()
	startWatcher()

	manager.Start()
	manager.StartNewEnvoy()
	manager.Wait()
	logger.Info("sucessfully terminated envoy manager and all envoys, so terminate hot-restarter")
}
