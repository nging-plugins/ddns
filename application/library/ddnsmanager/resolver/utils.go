package resolver

import (
	"net"
	"strings"

	"github.com/miekg/dns"
	"github.com/webx-top/echo/param"
)

func trimSpaceAndNotEmpty(s *string) bool {
	*s = strings.TrimSpace(*s)
	return len(*s) > 0
}

// ResolveDNS will query DNS for a given hostname.
// example: ResolveDNS(`webx.top`,`8.8.8.8`,`IPV4`)
func ResolveDNS(hostname, resolver, ipType string) (string, error) {
	var dnsType uint16
	if len(ipType) == 0 || strings.ToUpper(ipType) == `IPV4` {
		dnsType = dns.TypeA
	} else {
		dnsType = dns.TypeAAAA
	}
	var resolvers []string
	if len(resolver) > 0 {
		resolvers = param.Split(resolver, `,`).Filter(trimSpaceAndNotEmpty).Unique().String()
	}

	// If no DNS server is set in config file, falls back to default resolver.
	if len(resolvers) == 0 {
		dnsAdress, err := net.LookupHost(hostname)
		if err != nil {
			return "", err
		}

		return dnsAdress[0], nil
	}
	res := New(resolvers)
	// In case of i/o timeout
	res.RetryTimes = 5

	ip, err := res.LookupHost(hostname, dnsType)
	if err != nil {
		return "", err
	}

	if len(ip) > 0 {
		return ip[0].String(), nil
	}

	return ``, nil
}
