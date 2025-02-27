package app

type IApplication interface {
	AppName() string
	AppVersion() string
}
