package domain_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	alog "github.com/admpub/log"
	"github.com/nging-plugins/ddnsmanager/application/library/ddnsmanager"
	"github.com/nging-plugins/ddnsmanager/application/library/ddnsmanager/config"
	"github.com/nging-plugins/ddnsmanager/application/library/ddnsmanager/domain"
	"github.com/nging-plugins/ddnsmanager/application/library/ddnsmanager/domain/dnsdomain"
	"github.com/nging-plugins/ddnsmanager/application/library/ddnsmanager/interfaces"
	"github.com/nging-plugins/ddnsmanager/application/library/ddnsmanager/sender"
	"github.com/nging-plugins/ddnsmanager/application/library/ddnsretry"
	"github.com/stretchr/testify/assert"
	"github.com/webx-top/echo"
)

type testProvider struct {
	Domains  []*dnsdomain.Domain
	failMode bool
}

func (t *testProvider) Name() string {
	if t.failMode {
		return `test_fail`
	}
	return `test`
}

func (*testProvider) Description() string {
	return ``
}

func (*testProvider) SignUpURL() string {
	return ``
}

func (*testProvider) LineTypeURL() string {
	return ``
}

func (*testProvider) ConfigItems() echo.KVList {
	return nil
}

func (*testProvider) Support() dnsdomain.Support {
	return dnsdomain.Support{}
}

// Init 初始化
func (cf *testProvider) Init(settings echo.H, domains []*dnsdomain.Domain) error {
	cf.Domains = domains
	return nil
}

func (cf *testProvider) Update(ctx context.Context, recordType string, ipAddr string) error {
	time.Sleep(time.Second)
	if cf.failMode {
		maxIndex := len(cf.Domains) - 1
		for index, domain := range cf.Domains {
			log.Println(domain)
			if maxIndex == index {
				domain.UpdateStatus = dnsdomain.UpdatedFailed
				return errors.New(`test_fail: error`)
			}
			domain.UpdateStatus = dnsdomain.UpdatedSuccess
		}
	}
	for _, domain := range cf.Domains {
		log.Println(domain)
		domain.UpdateStatus = dnsdomain.UpdatedSuccess
	}
	return nil
}

func init() {
	ddnsmanager.Register(`test`, func() interfaces.Updater {
		return &testProvider{}
	})
	ddnsmanager.Register(`test_fail`, func() interfaces.Updater {
		return &testProvider{failMode: true}
	})
}

func makeDNSServiceConfig(provider string) *config.DNSService {
	return &config.DNSService{Enabled: true,
		Provider: provider,
		IPv4Domains: []*config.DNSDomain{
			{
				Domain: `test.webx.top`,
			},
			{
				Domain: `test2.webx.top`,
			},
			{
				Domain: `test3.webx.top`,
			},
		},
		IPv6Domains: []*config.DNSDomain{
			{
				Domain: `test6.webx.top`,
			},
			{
				Domain: `test61.webx.top`,
			},
			{
				Domain: `test62.webx.top`,
			},
		},
	}
}

func TestUpdate(t *testing.T) {
	defer alog.Close()
	c := &config.Config{
		NotifyMode: config.NotifyAll,
		IPv4: &config.NetIPConfig{
			Enabled: true,
			Type:    `netInterface`,
			NetInterface: &config.NetInterface{
				Name: `en0`,
			},
		},
		IPv6: &config.NetIPConfig{
			Enabled: true,
			Type:    `netInterface`,
			NetInterface: &config.NetInterface{
				Name: `en0`,
			},
		},
		DNSServices: []*config.DNSService{
			makeDNSServiceConfig(`test`),
			makeDNSServiceConfig(`test_fail`),
		},
	}
	d, err := domain.ParseDomain(c)
	assert.NoError(t, err)
	defer d.ClearIP()
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		time.Sleep(time.Second)
	}()
	sender.SetDebug(true)
	//*
	err = d.Update(ctx, c, false, `test`)
	assert.NoError(t, err)
	time.Sleep(2 * time.Second)
	//*/

	d.ClearIP()
	go func() {
		time.Sleep(5 * time.Second)
		cancel()
	}()
	err = d.Update(ctx, c, false, `test_fail`)
	assert.Error(t, ddnsretry.ErrCanceledRetry)
	time.Sleep(2 * time.Second)
}

func TestRetry(t *testing.T) {
	res := map[int]int{}
	for i := 1; i <= 10; i++ {
		res[i] = i
	}
	var errs []error
	wg := sync.WaitGroup{}
	chanErrors := make(chan error, len(res)+10)
	if true {
		for k := range res {
			wg.Add(1)
			go func(k int) {
				defer wg.Done()
				for j := 0; j < 5; j++ {
					time.Sleep(time.Second)
					fmt.Printf("[%d] %s\n", k, time.Now().Format(time.DateTime))
				}
				chanErrors <- fmt.Errorf(`test-error: %d`, k)
			}(k)
		}
	}
	wg.Wait()
	close(chanErrors)
	for chanErr := range chanErrors {
		if chanErr == nil {
			continue
		}
		t.Log(chanErr.Error())
		errs = append(errs, chanErr)
	}
}
