package ddnsmanager

import (
	"github.com/coscms/webcore/registry/navigate"

	"github.com/nging-plugins/ddnsmanager/application/handler"
)

var TopNavigate = handler.TopNavigate

func RegisterNavigate(nc *navigate.Collection) {
	nc.Backend.GetTop().AddChild(`tool`, -1, TopNavigate)
}
