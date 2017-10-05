package ip

import (
	"fmt"
	"github.com/ciena/spanneti/resolver"
)

func (p tenantIpPlugin) Event(key string, value interface{}) error {
	nets := p.GetRelatedTo(PLUGIN_NAME, key, value).([]TenantIpData)

	switch key {
	case "ip":
		fmt.Println("Event for IP:", value)

		ips := value.(TenantIP)

		if len(nets) == 1 {
			if err := p.trySetupIP(nets[0], ips); err != nil {
				return err
			}
		}

		if len(nets) == 0 {
			if err := p.tryCleanupIP(ips); err != nil {
				return err
			}
		}
	}
	return nil
}

//trySetupIP checks if the linkMap contains two containers, and if so, ensures interfaces are set up
func (p tenantIpPlugin) trySetupIP(data TenantIpData, ip TenantIP) error {
	fmt.Printf("Should link %s in %s\n",
		data.GetIfaceForIP(ip), data.containerId[0:12])

	containerPid, err := p.GetContainerPid(data.containerId)
	if err != nil {
		return err
	}

	if err := resolver.SetupTenantIpContainerLink(data.GetIfaceForIP(ip), containerPid, string(ip)); err != nil {
		return err
	}

	return p.requestSetup(ip)
}

//tryCleanupIP checks if the linkMap contains only one container, and if so, ensures interfaces are deleted
func (p tenantIpPlugin) tryCleanupIP(ip TenantIP) error {
	return p.requestDelete(ip)
}
