package main

import (
	"flag"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/sepaper/envoy-hot-restarter/internals/watcher"
	log "github.com/sirupsen/logrus"
)

var (
	logger           log.FieldLogger
	envoyConfigPath  string
	hotRestarterPath string
	envoyExecPath    string
	hotRestarter     *exec.Cmd
)

func init() {
	logger = log.New()
	log.SetLevel(log.InfoLevel)

	flag.StringVar(&hotRestarterPath, "hotRestarterPath", "hot-restarter.py", "Hot restarter python script path")
	flag.StringVar(&envoyConfigPath, "envoyConfigPath", "config.yaml", "Envoy config file path")
	flag.StringVar(&envoyExecPath, "envoyExecPath", "start_envoy.sh", "Envoy exec file path")
}

type EnvoyConfigFileEventHandler struct {
}

func (h *EnvoyConfigFileEventHandler) On_deleted() error {
	// send SIGTERM, SIGINT to python hot restarter
	return sendSigTermToHotRestarter()
}
func (h *EnvoyConfigFileEventHandler) On_created() error {
	// send SIGHUP to python hot restarter
	return sendSigHupToHotRestarter()
}
func (h *EnvoyConfigFileEventHandler) On_moved() error {
	// send SIGHUP to python hot restarter
	return sendSigHupToHotRestarter()
}
func (h *EnvoyConfigFileEventHandler) On_modified() error {
	// send SIGHUP to python hot restarter
	return sendSigHupToHotRestarter()
}

func sendSigTermToHotRestarter() error {
	return hotRestarter.Process.Signal(syscall.SIGTERM)
	/*
		p, _ := os.FindProcess(os.Getpid())
		p.Signal(syscall.SIGTERM)
	*/
}

func sendSigHupToHotRestarter() error {
	return hotRestarter.Process.Signal(syscall.SIGHUP)
}

func main() {
	pid := os.Getegid()
	ppid := os.Getppid()
	log.Info(pid, ppid)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	hotRestarter = exec.Command("python", hotRestarterPath, envoyExecPath)
	err := hotRestarter.Start()
	if err != nil {
		log.Error("")
		return
	}

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGCHLD)

	//signal handler
	go func() {
		for {
			select {
			case sig := <-sigs:
				if sig == syscall.SIGINT || sig == syscall.SIGTERM {
					log.Info("")
					sendSigTermToHotRestarter()
				} else if sig == syscall.SIGCHLD {
					log.Info("")
					done <- true
				}
			}
		}
	}()

	// watcher for file change
	go func() {
		watcher.Watch(envoyConfigPath, &EnvoyConfigFileEventHandler{})
	}()

	log.Info("")
	<-done
}
