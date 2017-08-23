# Spanneti
*Span*s *net*works, creating a *spa*ghet*ti* of connections.

## Description
Spanneti listens for specially-tagged containers, and builds container networks based on those tags.
It was specifically designed to set up VNF chains, so it currently supports point-to-point L2 links (on a single host, or between multiple hosts),
and OLT setup (s-tag/c-tag interfaces for the currently-limited tag-forwarding of ONOS).


## HA
Spanneti was designed with HA & fault tolerance as the highest priority, so it can crash, be restarted, or (in theory) upgraded, without disrupting traffic.


## Design
It currently ignores the docker/k8s plugin system (our use case was not supported by the docker plugin API.)

It runs as a single container (13MB total) per compute, and does not need to maintain state.  (on restart, will diff container tags vs current network)
