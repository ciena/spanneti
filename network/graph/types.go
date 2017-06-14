package graph

type LinkID string
type ContainerID string

type ContainerNetwork struct {
	//interface - UUID mapping
	Links       map[string]LinkID `json:"links,omitempty"`
	OLT         oltLink           `json:"olt"`
	IP          string            `json:"ip"`
	ContainerId ContainerID       `json:"-"`
}

type oltLink struct {
	STag uint16
	CTag uint16
}
