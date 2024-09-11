package sender

import "github.com/coscms/webcore/registry/alert"

func init() {
	alert.Topics.Add(`ddnsUpdate`, `DDNS更新`)
}
