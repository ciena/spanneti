# Spanneti
**Span**s **net**works, creating a **spa**gh**etti** of connections.


## Description
Spanneti listens for specially-tagged containers, and builds container networks based on those tags.
It was specifically designed to set up VNF chains, so it currently supports point-to-point L2 links (on a single host, or between multiple hosts),
and OLT setup (s-tag/c-tag interfaces for the currently-limited tag-forwarding of ONOS).

## Required Environment
Since the Spanneti only supports kubernetes platform now, this install document will focus on kubernetes platform.
1. You should prepare a kubernetes cluster, you can use kubernates build-in tool [kubeadm](https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm/) to setup cluster
or follow some userful documents such as [kubernetes-the-hard-way](https://github.com/kelseyhightower/kubernetes-the-hard-way) step by step.
2. In order to make kubernetes works, don't forget to setup a network for your kubernetes cluster, you can use Container Network Plugin(CNI) based networks, such as Flannel, Weave Net and so on.
3. Make sure everthing goes well, and you can start to install spanneti now.


## Running
Deploy in a k8s cluster using the k8s deployment file.
`kubectl apply -f k8s.yml`
(You may want to pin the version)


### Configuration
Environment variables can be changed in the k8s deployment file.
 * DNS_NAME - name of the dns entry to lookup for peer discovery
 * HOST_INTERFACE_NAME - name of the interface on which to setup networking
