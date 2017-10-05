package ip

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (plugin *tenantIpPlugin) newHttpHandler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/resync", plugin.resyncHandler).Methods(http.MethodPost)
	return r
}

func (plugin *tenantIpPlugin) resyncHandler(w http.ResponseWriter, r *http.Request) {
	nets := plugin.GetAllDataFor(PLUGIN_NAME).([]TenantIpData)

	go func() {
		for _, net := range nets {
			for _, tenantIp := range net.IP {
				plugin.FireEvent(PLUGIN_NAME, "ip", tenantIp)
			}
		}
	}()

	w.WriteHeader(http.StatusAccepted)
}
