# Spanneti
**Span**s **net**works, creating a **spa**gh**etti** of connections.


## Description
Spanneti listens for specially-tagged containers, and builds container networks based on those tags.
It was specifically designed to set up VNF chains, so it currently supports point-to-point L2 links (on a single host, or between multiple hosts),
and OLT setup (s-tag/c-tag interfaces for the currently-limited tag-forwarding of ONOS).


## Running
Deploy in a k8s cluster using the k8s deployment file.
`kubectl apply -f k8s.yml`
(You may want to pin the version)


### Configuration
Environment variables can be changed in the k8s deployment file.
 * DNS_NAME - name of the dns entry to lookup for peer discovery
 * HOST_INTERFACE_NAME - name of the interface on which to setup networking
