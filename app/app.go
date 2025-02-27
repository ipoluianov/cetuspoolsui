package app

import (
	"fmt"

	"github.com/ipoluianov/gomisc/logger"
)

func Init(name string, displayName string, startFunc func(), stopFunc func()) {
	Name = name
	ServiceName = name
	ServiceDisplayName = displayName
	ServiceDescription = ""
	ServiceRunFunc = RunAsService
	ServiceStopFunc = StopService
	AppStartFunc = startFunc
	AppStopFunc = stopFunc
}

func Start() {
	TuneFDs()
	AppStartFunc()
}

func Stop() {
	AppStopFunc()
}

func RunDesktop() {
	logger.Println("Running as console application")
	Start()
	fmt.Scanln()
	logger.Println("Console application exit")
}

func RunAsService() error {
	Start()
	return nil
}

func StopService() {
	Stop()
}
