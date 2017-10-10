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


## Install
Deploy in a kubernetes cluster using the kubernetes deployment file.

#### Apply Spanneti to kubernetes cluster
Type `kubectl apply -f k8s.yml` on master node to deploy spanneti.
(You may want to pin the version)

After deploy, you can use following commands to check status of spanneti.

#### Check Spanneti Service
Type `kubectl get services` and the result will look like below
```
NAME         CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
kubernetes   10.96.0.1    <none>        443/TCP    9d
spanneti     None         <none>        8080/TCP   9d
```

#### Check Spanneti Pods
Type `kubectl get pods` and the result will look like below (Assume there're two nodes in your kubernetes cluster)
```
NAME                        READY     STATUS    RESTARTS   AGE
spanneti-b6bc3              1/1       Running   4          9d
spanneti-zz0rz              1/1       Running   5          9d
```

Make sure the status of each spanneti pods is `Running` before we start to conntect containers via Spanneti service.

## Usage
As mentioned above, Spanneti build container networks based on tags.  
In order to make container belongs to same networks, you should set meta data to container via `--label com.opencord.network.graph={tag format}` when you run it.

#### tag format
```
{"olt": {
    "<ethName>": {
        "s-tag":<sTag>,
        "c-tag":<cTag>
    }},
    ...
"links": {
    "<ethName2>":"<link_ID>",
    ...
}}
```

## Example
In this example, you will run two containers with same tag and Spanneti will setup a point-to-point (L2) link between those two container on the same host.

#### Watch Spanneti log
Type `kubtctl log -f spanneti-xxxx` to check specific Spanneti Pod log.


#### Create containers
Run following command on any node which runs a Spanneti Pod.  
`docker run --name=e1 -d --restart=always --label=com.opencord.network.graph=\{\"links\":\{\"iface0\":\"UUID-1\"\}\} --net=none hwchiu/netutils sleep 100000`  
`docker run --name=e2 -d --restart=always --label=com.opencord.network.graph=\{\"links\":\{\"iface0\":\"UUID-1\"\}\} --net=none hwchiu/netutils sleep 100000`  
Now. Spanneti will connect that two containers (e1,e2) together because they has the same tag (**UUID-1**)  

#### Testing
You can use following command to verify point-to-point network.

Type `docker exec -it e1 ping 8.8.8.8 -I iface0`, the container e1 will send a packet via its iface interface.  
Type `docker exec -it e2 tcpdump -i iface0` and you will see ARP request issued from container e1.

If you want to learn more about Spanneti example, you can refer to this [video](https://youtu.be/U46WBzygD7s?t=17m44s).

## Configuration
Environment variables can be changed in the k8s deployment file.
 * DNS_NAME - name of the dns entry to lookup for peer discovery
 * HOST_INTERFACE_NAME - name of the interface on which to setup networking
