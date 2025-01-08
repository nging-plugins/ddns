package handler

import (
	"github.com/coscms/webcore/library/navigate"
	"github.com/webx-top/echo"
)

var TopNavigate = &navigate.Item{
	Display: true,
	Name:    echo.T(`DDNS`),
	Action:  `ddns`,
}
