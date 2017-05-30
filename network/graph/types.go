package graph

type LinkID string
type ContainerID string

type ContainerNetwork struct {
	//interface - UUID mapping
	Links       map[string]LinkID `json:"links,omitempty"`
	ContainerId ContainerID       `json:"-"`
}