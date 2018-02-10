package ip2proxy

import "encoding/binary"

// DbType is the type of db
type DbType uint8

const (
	// UnknownDbType is the type for unknown db type
	UnknownDbType DbType = 0
	// PX1 is the IP2Proxy IP-COUNTRY database
	PX1 DbType = 1
	// PX2 is the IP2Proxy IP-PROXYTYPE-COUNTRY database
	PX2 DbType = 2
	// PX3 is the IP2Proxy IP-PROXYTYPE-COUNTRY-REGION-CITY database
	PX3 DbType = 3
	// PX4 is the P2Proxy IP-PROXYTYPE-COUNTRY-REGION-CITY-ISP database
	PX4 DbType = 4
)


// ProxyType is the type of proxy detected
type ProxyType uint8

const (
	// ProxyNA are hosts which proxy status is not available (db without the info or invalid data).
	ProxyNA ProxyType = iota
	// ProxyNOT are hosts not detected as proxy.
	ProxyNOT
	// ProxyVPN are Anonymizing VPN services. These services offer users a publicly accessible VPN for the purpose of
	// hiding their IP address.
	ProxyVPN
	// ProxyTOR are Tor Exit Nodes. The Tor Project is an open network used by those who wish to maintain anonymity.
	ProxyTOR
	// ProxyDCH Are Hosting Provider, Data Center or Content Delivery Network. Since hosting providers and data centers
	// can serve to provide anonymity, the Anonymous IP database flags IP addresses associated with them.
	ProxyDCH
	// ProxyPUB are Public Proxies. These are services which make connection requests on a user's behalf.
	// Proxy server software can be configured by the administrator to listen on some specified port.
	// These differ from VPNs in that the proxies usually have limited functions compare to VPNs.
	ProxyPUB
	// ProxyWEB are Web Proxies. These are web services which make web requests on a user's behalf.
	// These differ from VPNs or Public Proxies in that they are simple web-based proxies rather than operating at the IP address and other ports level.
	ProxyWEB
)

// get proxy type according to name
func proxyNameToProxyType(name string) ProxyType {
	switch name {
	case "-":
		return ProxyNOT
	case "VPN":
		return ProxyVPN
	case "TOR":
		return ProxyTOR
	case "DCH":
		return ProxyDCH
	case "PUB":
		return ProxyPUB
	case "WEB":
		return ProxyWEB
	default:
		return ProxyNA
	}
}

// Fields indexes.
var countryPos = []uint8{0, 2, 3, 3, 3}
var regionPos = []uint8{0, 0, 0, 4, 4}
var cityPos = []uint8{0, 0, 0, 5, 5}
var ispPos = []uint8{0, 0, 0, 0, 6}
var proxytypePos = []uint8{0, 0, 2, 2, 2}

// File endianness
var fileEndianness = binary.LittleEndian

// Maximum index count
const maxIndexes = 65536

