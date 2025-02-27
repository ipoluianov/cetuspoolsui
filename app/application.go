package app

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kardianos/osext"
	"github.com/kardianos/service"
)

var Name string
var Description string
var ServiceName string
var ServiceDisplayName string
var ServiceDescription string
var ServiceRunFunc func() error
var ServiceStopFunc func()

var AppStartFunc func()
var AppStopFunc func()

type Application struct {
	Name    string
	Version string
}

var App Application

func SetAppPath() {
	exePath, _ := osext.ExecutableFolder()
	os.Chdir(exePath)

}

func init() {
	SetAppPath()
}

func TryService() bool {
	serviceFlagPtr := flag.Bool("service", false, "Run as service")
	installFlagPtr := flag.Bool("install", false, "Install service")
	uninstallFlagPtr := flag.Bool("uninstall", false, "Uninstall service")
	startFlagPtr := flag.Bool("start", false, "Start service")
	stopFlagPtr := flag.Bool("stop", false, "Stop service")

	flag.Parse()

	if *serviceFlagPtr {
		runService()
		return true
	}

	if *installFlagPtr {
		InstallService()
		return true
	}

	if *uninstallFlagPtr {
		UninstallService()
		return true
	}

	if *startFlagPtr {
		StartServiceBase()
		return true
	}

	if *stopFlagPtr {
		StopServiceBase()
		return true
	}

	return false
}

func NewSvcConfig() *service.Config {
	var SvcConfig = &service.Config{
		Name:        ServiceName,
		DisplayName: ServiceDisplayName,
		Description: ServiceDescription,
	}
	SvcConfig.Arguments = append(SvcConfig.Arguments, "-service")
	return SvcConfig
}

func InstallService() {
	fmt.Println("Service installing")
	prg := &program{}
	s, err := service.New(prg, NewSvcConfig())
	if err != nil {
		log.Fatal(err)
	}
	s.Install()
	fmt.Println("Service installed")
}

func UninstallService() {
	fmt.Println("Service uninstalling")
	prg := &program{}
	s, err := service.New(prg, NewSvcConfig())
	if err != nil {
		log.Fatal(err)
	}
	s.Uninstall()
	fmt.Println("Service uninstalled")
}

func StartServiceBase() {
	fmt.Println("Service starting")
	prg := &program{}
	s, err := service.New(prg, NewSvcConfig())
	if err != nil {
		log.Fatal(err)
	}
	err = s.Start()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Service started")
	}
}

func StopServiceBase() {
	fmt.Println("Service stopping")
	prg := &program{}
	s, err := service.New(prg, NewSvcConfig())
	if err != nil {
		log.Fatal(err)
	}
	err = s.Stop()
	if err != nil {
		log.Fatal(err)
		return
	} else {
		fmt.Println("Service stopped")
	}
}

func runService() {
	prg := &program{}
	s, err := service.New(prg, NewSvcConfig())
	if err != nil {
		log.Fatal(err)
	}
	serviceLogger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}
	err = s.Run()
	if err != nil {
		serviceLogger.Error(err)
	}
}

var serviceLogger service.Logger

type program struct{}

func (p *program) Start(s service.Service) error {
	return ServiceRunFunc()
}

func (p *program) Stop(s service.Service) error {
	ServiceStopFunc()
	return nil
}
