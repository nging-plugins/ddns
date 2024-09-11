package ddnsmanager

import (
	"github.com/coscms/webcore/library/module"

	"github.com/nging-plugins/ddnsmanager/application/handler"
)

const ID = `ddns`

var Module = module.Module{
	TemplatePath: map[string]string{
		ID: `ddnsmanager/template/backend`,
	},
	AssetsPath:  []string{},
	Navigate:    RegisterNavigate,
	Route:       handler.RegisterRoute,
	DBSchemaVer: 0.0000,
}
