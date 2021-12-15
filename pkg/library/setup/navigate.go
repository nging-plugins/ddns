package setup

import (
	"github.com/admpub/nging/v4/application/handler/tool"
	"github.com/admpub/nging/v4/application/registry/navigate"
)

func init() {
	tool.TopNavigate.Add(-1, &navigate.Item{
		Display: true,
		Name:    `DDNS`,
		Action:  `ddns`,
	})
}
