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


### Configuration
Environment variables can be changed in the k8s deployment file.
 * DNS_NAME - name of the dns entry to lookup for peer discovery
 * HOST_INTERFACE_NAME - name of the interface on which to setup networking
