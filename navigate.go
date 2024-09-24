package ddnsmanager

import (
	"github.com/coscms/webcore/library/module"
	"github.com/nging-plugins/ddnsmanager/application/handler"
)

var TopNavigate = handler.TopNavigate

func RegisterNavigate(nc module.Navigate) {
	nc.Backend().GetTop().AddChild(`tool`, -1, TopNavigate)
}
