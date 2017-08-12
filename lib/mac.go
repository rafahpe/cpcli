package lib

import (
	"strings"
)

// MAC address
type MAC string

// NewMAC builds a MAC address, uppercased.
func NewMAC(mac string) MAC {
	for _, decorator := range []string{":", "-", "."} {
		mac = strings.Replace(mac, decorator, "", -1)
	}
	mac = "000000000000" + strings.TrimSpace(strings.ToUpper(mac))
	siz := len(mac)
	return MAC(mac[siz-12 : siz])
}

func (mac MAC) split(size int) []string {
	result := make([]string, 0, (12/size)+1)
	for i := 0; i < 12; i += size {
		result = append(result, string(mac[i:i+size]))
	}
	return result
}

// Colon returns the uppercased, colon-separated representation of the MAC
func (mac MAC) Colon() string {
	return strings.Join(mac.split(2), ":")
}

// Hyphen returns the uppercased, hyphen-separated representation of the MAC
func (mac MAC) Hyphen() string {
	return strings.Join(mac.split(2), "-")
}

// Dot returns the uppercased, dot-separated representation of the MAC
func (mac MAC) Dot() string {
	return strings.Join(mac.split(4), ".")
}
