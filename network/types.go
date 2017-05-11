package network

type LinkID string
type ContainerID string


type ContainerNetwork struct {
	//interface - UUID mapping
	Links map[string]LinkID `json:"links,omitempty"`

	containerId ContainerID `json:"-"`
}

func (contNet ContainerNetwork)getIfaceFor(linkId LinkID) string{
	for iface, id:=range contNet.Links{
		if id == linkId{
			return iface
		}
	}
	panic("linkId not found")
}