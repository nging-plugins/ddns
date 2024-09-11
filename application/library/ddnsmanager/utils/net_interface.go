package utils

import (
	"fmt"
	"net"

	"github.com/admpub/log"
	"github.com/admpub/resty/v2"

	"github.com/coscms/webcore/library/ip2region"
	"github.com/coscms/webcore/library/restclient"
	"github.com/nging-plugins/ddnsmanager/application/library/ddnsmanager/config"
)

// NetInterface 本机网络
type NetInterface struct {
	Name    string
	Address []string
}

var ipv6Unicast *net.IPNet

func init() {
	var err error
	// https://en.wikipedia.org/wiki/IPv6_address#General_allocation
	_, ipv6Unicast, err = net.ParseCIDR("2000::/3")
	if err != nil {
		panic(err)
	}
}

// GetNetInterface 获得网卡地址 (返回ipv4, ipv6地址)
func GetNetInterface(interfaceName string) (ipv4NetInterfaces []NetInterface, ipv6NetInterfaces []NetInterface, err error) {
	var allNetInterfaces []net.Interface
	if len(interfaceName) > 0 {
		var ifaces *net.Interface
		ifaces, err = net.InterfaceByName(interfaceName)
		if err == nil {
			allNetInterfaces = append(allNetInterfaces, *ifaces)
		}
	} else {
		allNetInterfaces, err = net.Interfaces()
	}
	if err != nil {
		log.Error("net.Interfaces failed, err: ", err.Error())
		return ipv4NetInterfaces, ipv6NetInterfaces, err
	}

	for _, netInter := range allNetInterfaces {
		if (netInter.Flags & net.FlagUp) == 0 {
			continue
		}
		addrs, _ := netInter.Addrs()
		var ipv4 []string
		var ipv6 []string

		for _, address := range addrs {
			ipnet, ok := address.(*net.IPNet)
			if !ok || !ipnet.IP.IsGlobalUnicast() {
				continue
			}
			// 需匹配全局单播地址
			_, bits := ipnet.Mask.Size()
			switch bits {
			case 128:
				if ipv6Unicast.Contains(ipnet.IP) {
					ipv6 = append(ipv6, ipnet.IP.String())
				}
			case 32:
				ipv4 = append(ipv4, ipnet.IP.String())
			}
		}

		if len(ipv4) > 0 {
			ipv4NetInterfaces = append(
				ipv4NetInterfaces,
				NetInterface{
					Name:    netInter.Name,
					Address: ipv4,
				},
			)
		}

		if len(ipv6) > 0 {
			ipv6NetInterfaces = append(
				ipv6NetInterfaces,
				NetInterface{
					Name:    netInter.Name,
					Address: ipv6,
				},
			)
		}
	}

	return ipv4NetInterfaces, ipv6NetInterfaces, nil
}

// GetIPv4Addr 获得IPv4地址
func GetIPv4Addr(conf *config.NetIPConfig) (result string, err error) {
	// 判断从哪里获取IP
	switch conf.Type {
	case "netInterface":
		// 从网卡获取IP
		var ipv4 []NetInterface
		ipv4, _, err = GetNetInterface(conf.NetInterface.Name)
		if err != nil {
			err = fmt.Errorf("从网卡获得IPv4失败: %w", err)
			return
		}

		for _, netInterface := range ipv4 {
			if netInterface.Name != conf.NetInterface.Name || len(netInterface.Address) == 0 {
				continue
			}
			if conf.NetInterface.Filter == nil {
				result = netInterface.Address[0]
				return
			}
			for _, addr := range netInterface.Address {
				if conf.NetInterface.Filter.Match(addr) {
					result = addr
					return
				}
			}
		}

		err = fmt.Errorf("从网卡中获得IPv4失败! 网卡名: %s", conf.NetInterface.Name)
		return
	case "cmd":
		var _result []byte
		_result, err = conf.CommandLine.Exec()
		if err != nil {
			err = fmt.Errorf("读取IPv4结果失败: %w", err)
			return
		}
		result = ip2region.FindIPv4(string(_result))
		return
	default:
		if len(conf.NetIPApiUrl) == 0 {
			var wanIP ip2region.WANIP
			wanIP, err = ip2region.GetWANIP(0, 4)
			if err != nil {
				err = fmt.Errorf("读取IPv4结果失败: %w", err)
				return
			}
			result = wanIP.IP
			return
		}
		client := restclient.Resty()
		var resp *resty.Response
		resp, err = client.Get(conf.NetIPApiUrl)
		if err != nil {
			err = fmt.Errorf("未能获得IPv4地址: %w 查询URL: %s", err, conf.NetIPApiUrl)
			return
		}
		if !resp.IsSuccess() {
			err = fmt.Errorf("未能获得IPv4地址: [%d]%s 查询URL: %s", resp.StatusCode(), resp.Status(), conf.NetIPApiUrl)
			return
		}
		body := resp.Body()
		if len(body) == 0 {
			err = fmt.Errorf("读取IPv4结果失败: API 接口响应内容为空 查询URL: %s", conf.NetIPApiUrl)
			return
		}
		result = ip2region.FindIPv4(string(body))
		return
	}
}

// GetIPv6Addr 获得IPv6地址
func GetIPv6Addr(conf *config.NetIPConfig) (result string, err error) {
	// 判断从哪里获取IP
	switch conf.Type {
	case "netInterface":
		// 从网卡获取IP
		var ipv6 []NetInterface
		_, ipv6, err = GetNetInterface(conf.NetInterface.Name)
		if err != nil {
			err = fmt.Errorf("从网卡获得IPv6失败: %w", err)
			return
		}

		for _, netInterface := range ipv6 {
			if netInterface.Name != conf.NetInterface.Name || len(netInterface.Address) == 0 {
				continue
			}
			if conf.NetInterface.Filter == nil {
				result = netInterface.Address[0]
				return
			}
			for _, addr := range netInterface.Address {
				if conf.NetInterface.Filter.Match(addr) {
					result = addr
					return
				}
			}
		}

		log.Error("从网卡中获得IPv6失败! 网卡名: ", conf.NetInterface.Name)
		return
	case "cmd":
		var _result []byte
		_result, err = conf.CommandLine.Exec()
		if err != nil {
			err = fmt.Errorf("读取IPv6结果失败: %w", err)
			return
		}
		result = ip2region.FindIPv6(string(_result))
		return
	default:
		if len(conf.NetIPApiUrl) == 0 {
			var wanIP ip2region.WANIP
			wanIP, err = ip2region.GetWANIP(0, 6)
			if err != nil {
				err = fmt.Errorf("读取IPv6结果失败: %w", err)
				return
			}
			result = wanIP.IP
			return
		}

		client := restclient.Resty()
		var resp *resty.Response
		resp, err = client.Get(conf.NetIPApiUrl)
		if err != nil {
			err = fmt.Errorf("未能获得IPv6地址: %w 查询URL: %s", err, conf.NetIPApiUrl)
			return
		}
		if !resp.IsSuccess() {
			err = fmt.Errorf("未能获得IPv6地址: [%d]%s 查询URL: %s", resp.StatusCode(), resp.Status(), conf.NetIPApiUrl)
			return
		}
		body := resp.Body()
		if len(body) == 0 {
			err = fmt.Errorf("读取IPv6结果失败: API 接口响应内容为空 查询URL: %s", conf.NetIPApiUrl)
			return
		}
		result = ip2region.FindIPv6(string(body))
		return
	}
}
