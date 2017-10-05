package ip

import (
	"github.com/ciena/spanneti/spanneti/graph"
)

type TenantIP string

type TenantIpData struct {
	IP          map[string]TenantIP `json:"ip" spanneti:"ip"`
	containerId graph.ContainerID
}

func (data TenantIpData) GetIfaceForIP(tenantIp TenantIP) string {
	for iface, ip := range data.IP {
		if ip == tenantIp {
			return iface
		}
	}
	panic("linkId not found")
}

func (data TenantIpData) SetContainerID(containerId graph.ContainerID) graph.PluginData {
	data.containerId = containerId
	return data
}
