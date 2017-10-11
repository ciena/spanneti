# Spanneti
**Span**s **net**works, creating a **spa**gh**etti** of connections.


## Description
Spanneti listens for specially-tagged containers, and builds container networks based on those tags.
It was specifically designed to set up VNF chains, so it currently supports point-to-point L2 links (on a single host, or between multiple hosts),
and OLT setup (s-tag/c-tag interfaces for the currently-limited tag-forwarding of ONOS).

## Environment Setup
You will likely want a functioning container management system.
For now, k8s is used as the reference implementation.
Documentation for this is available elsewhere, try [kubeadm](https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm/) or [kubernetes-the-hard-way](https://github.com/kelseyhightower/kubernetes-the-hard-way)
In addition, you will need:
* A pre-defined `fabric` interface on each host, which spanneti will use when creating networks.
* A DNS entry which spanneti can use for peer discovery.  In the `k8s.yml`, this is done for you, by creating a `Service`.

## Install

#### Apply Spanneti to K8s Cluster
Type `kubectl apply -f k8s.yml` on master node to deploy spanneti.
(You may want to pin the version)

After deploy, you can use following commands to check status of spanneti.

#### Check Spanneti Service
`kubectl get services`

The result should look like:
```
NAME         CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
kubernetes   10.96.0.1    <none>        443/TCP    9d
spanneti     None         <none>        8080/TCP   9d
```

#### Check Spanneti Pods
`kubectl get pods`

The result should look something like: (with one entry per node in the cluster)
```
NAME                        READY     STATUS    RESTARTS   AGE
spanneti-b6bc3              1/1       Running   4          9d
spanneti-zz0rz              1/1       Running   5          9d
```

The status of each spanneti pods should be `Running`.

TODO: Deploy in a swarm cluster using a `docker-compose.yml` file.

## Usage
As mentioned above, spanneti builds container networks based on tags.  
In order to network containers, metadata will need to be added by specifying:
* a container label `com.opencord.network.graph={tag-format}`, or
* an environment variable `OPENCORD_NETWORK_GRAPH={tag-format}`

#### Tag Format
```
{"olt": {
    "<ethName>": {
        "s-tag":<sTag>,
        "c-tag":<cTag>
    },
    ...
  },
  "links": {
    "<ethName2>":"<link_ID>",
    ...
}}
```
(`olt` section is implemented by the `olt` plugin, `links` section is implemented by the `link` plugin)

## Basic Example
In this example, we will run two containers, and spanneti will setup a point-to-point (L2) link between those two containers.

This example was created from this [demo video](https://youtu.be/U46WBzygD7s?t=17m44s).

#### Watch spanneti log
Type `kubtctl log -f spanneti-xxxx` to follow a spanneti instance's logs.


#### Create containers
Run following commands on any nodes which run a spanneti Pod. 
`docker run --name=e1 -d --restart=always --label=com.opencord.network.graph=\{\"links\":\{\"iface0\":\"UUID-1\"\}\} --net=none hwchiu/netutils sleep 100000`  
`docker run --name=e2 -d --restart=always --label=com.opencord.network.graph=\{\"links\":\{\"iface0\":\"UUID-1\"\}\} --net=none hwchiu/netutils sleep 100000`  
Now. spanneti will connect that two containers (e1,e2) together because they has the same tag (**UUID-1**) 

#### Testing
You can use following command to verify point-to-point network.

`docker exec -it e1 ping 8.8.8.8 -I iface0`, the container e1 will send packets via its iface interface.  
`docker exec -it e2 tcpdump -i iface0` and you will see ARP request issued from container e1.

Note that the ping will not get replies, as there is no DHCP service running in these containers.

## Configuration
Environment variables can be changed in the k8s deployment file.
 * DNS_NAME - name of the dns entry to lookup for peer discovery
 * HOST_INTERFACE_NAME - name of the interface on which to setup networking
