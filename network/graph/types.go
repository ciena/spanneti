package graph

import "fmt"

type LinkID string
type ContainerID string

type ContainerNetwork struct {
	//interface - UUID mapping
	Links       map[string]LinkID  `json:"links,omitempty"`
	OLT         map[string]OltLink `json:"olt"`
	IP          string             `json:"ip"`
	ContainerId ContainerID        `json:"-"`
}

type OltLink struct {
	STag uint16 `json:"s-tag"`
	CTag uint16 `json:"c-tag"`
}

func (olt OltLink) String() string {
	return fmt.Sprintf("%d-%d", olt.STag, olt.CTag)
}
