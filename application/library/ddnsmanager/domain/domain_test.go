package domain

import (
	"testing"

	"github.com/nging-plugins/ddnsmanager/application/library/ddnsmanager/config"
	"github.com/nging-plugins/ddnsmanager/application/library/ddnsmanager/domain/dnsdomain"
	"github.com/stretchr/testify/assert"
	"github.com/webx-top/com"
	"github.com/webx-top/echo/param"
)

func TestDomain(t *testing.T) {
	domains, err := parseDomainArr([]*config.DNSDomain{
		{Domain: `a.b.c.test.com.cn`},
		{Domain: `w.webx.top`},
		{Domain: `dl.eget.io`},
		{Domain: `webx.top`},
	})
	if err != nil {
		panic(err)
	}
	com.Dump(domains)
	expected := []*dnsdomain.Domain{
		{
			DomainName:   "test.com.cn",
			SubDomain:    "a.b.c",
			UpdateStatus: "",
			Extra:        param.Store{},
		},
		{
			DomainName:   "webx.top",
			SubDomain:    "w",
			UpdateStatus: "",
			Extra:        param.Store{},
		},
		{
			DomainName:   "eget.io",
			SubDomain:    "dl",
			UpdateStatus: "",
			Extra:        param.Store{},
		},
		{
			DomainName:   "webx.top",
			SubDomain:    "",
			UpdateStatus: "",
			Extra:        param.Store{},
		},
	}
	assert.Equal(t, expected, domains)
}
