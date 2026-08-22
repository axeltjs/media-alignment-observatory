package apihandler

import "github.com/hafidluqman50/maoi/src/service"

var svcs service.Registry

type Services struct {
	Registry service.Registry
}

func ConfigureServices(s Services) {
	svcs = s.Registry
}
