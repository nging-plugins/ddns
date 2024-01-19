package resolver

import (
	"fmt"
	"testing"
)

func TestResolveDNS(t *testing.T) {
	ip, err := ResolveDNS(`www.webx.top`, `8.8.8.8`, `IPV4`)
	if err != nil {
		panic(err)
	}
	fmt.Println(`ResolveDNS:`, ip)
	ip, err = ResolveDNS(`ip6.enote.vip`, `8.8.8.8`, `IPV6`)
	if err != nil {
		panic(err)
	}
	fmt.Println(`ResolveDNS:`, ip)
	ip, err = ResolveDNS(`ip6.enote.vip`, ``, `IPV6`)
	if err != nil {
		panic(err)
	}
	fmt.Println(`ResolveDNS:`, ip)
}
