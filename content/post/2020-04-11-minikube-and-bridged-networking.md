---
title: Minikube and Bridged Networking
description: Exposing a VirtualBox Kubernetes cluster to the local network.
date: 2020-04-11
tags:
- bridged-networking
- kubernetes
- macos
- minikube
- networking
- virtualbox
aliases:
- /blog/2020/04/11/bridged-networking-with-minikube.html
- /blog/2020/04/11/minikube-and-bridged-networking.html
---

This week I was running some experiments with [Kubernetes](https://kubernetes.io/). Typically I use [minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/) locally, but, having just cleaned off an old Mac, I wanted to try hosting the cluster there instead. This was a bit of a challenge since minikube is focused on local development -- the Kubernetes cluster it creates is not accessible outside the machine where it is running. Eventually, after some trial and error, I was able to automate the setup so that any device on my local network could access the development cluster and its services.

To start, I need to make sure minikube and [VirtualBox](https://virtualbox.org/) are installed.

{{< snippet dir="appendix/2020-04-11-minikube-and-bridged-networking" file="minikube.sh" lines="57" stripprefix="  " lang="bash" >}}

With nothing else running on the old Mac, I want Kubernetes to have all the available resources. I use the [`sysctl` tool](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man3/sysctl.3.html) to check the available CPUs and memory on the system and configure minikube to use it.

{{< snippet dir="appendix/2020-04-11-minikube-and-bridged-networking" file="minikube.sh" lines="75-80" stripprefix="  " lang="bash" >}}

Once minikube finishes provisioning the virtual machine, I reconfigure it with an additional bridged network. Using the [`VBoxManage` CLI](https://www.virtualbox.org/manual/ch08.html) I create a new NIC on the next empty interface, bridge it to the Mac's default network interface, and, optionally, use a specific MAC address to support DHCP reservations.

{{< snippet dir="appendix/2020-04-11-minikube-and-bridged-networking" file="minikube.sh" lines="88-100" stripprefix="  " lang="bash" >}}

After the VM is restarted, I can use `minikube ssh` to verify it has received an IP on my local network (e.g. `192.0.2.113`).

{{< snippet dir="appendix/2020-04-11-minikube-and-bridged-networking" file="minikube.sh" lines="15" stripprefix="  " lang="bash" >}}

Now that it has an IP on my local network, I want to actually use it. However, the certificates that Kubernetes has are only signed for the host-only IP address (i.e. `192.168.99.106`). Rather than trying to regenerate those certificates, I decided to use one of the internal domains from the certificate -- `kubernetes.default.svc.cluster.local`. After I update my network's DNS server (or `/etc/hosts` file), I update `$KUBECONFIG` to use the hostname, too.

{{< snippet dir="appendix/2020-04-11-minikube-and-bridged-networking" file="minikube.sh" lines="114-115" stripprefix="  " lang="bash" >}}

With everything working locally from the old Mac, it's time to get a `$KUBECONFIG` for accessing it from my laptop.

{{< snippet dir="appendix/2020-04-11-minikube-and-bridged-networking" file="minikube.sh" lines="19" stripprefix="  " lang="bash" >}}

Now, with a copy of `$KUBECONFIG` on my laptop I'm able to deploy something to my remote, minikube Kubernetes cluster. For example, the [Dashboard UI](https://kubernetes.io/docs/tasks/access-application-cluster/web-ui-dashboard/). After all the resources are ready, I can use [`kubectl proxy`](https://kubernetes.io/docs/tasks/access-kubernetes-api/http-proxy-access-api/) to see it up and running.

```console
$ kubectl cluster-info
Kubernetes master is running at https://kubernetes.default.svc:8443
KubeDNS is running at https://kubernetes.default.svc:8443/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy

$ kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.0.0-rc7/aio/deploy/recommended.yaml
namespace/kubernetes-dashboard created
serviceaccount/kubernetes-dashboard created
service/kubernetes-dashboard created
secret/kubernetes-dashboard-certs created
secret/kubernetes-dashboard-csrf created
secret/kubernetes-dashboard-key-holder created
configmap/kubernetes-dashboard-settings created
role.rbac.authorization.k8s.io/kubernetes-dashboard created
clusterrole.rbac.authorization.k8s.io/kubernetes-dashboard created
rolebinding.rbac.authorization.k8s.io/kubernetes-dashboard created
clusterrolebinding.rbac.authorization.k8s.io/kubernetes-dashboard created
deployment.apps/kubernetes-dashboard created
service/dashboard-metrics-scraper created
deployment.apps/dashboard-metrics-scraper created

$ kubectl -n kubernetes-dashboard get service/kubernetes-dashboard
NAME                   TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
kubernetes-dashboard   ClusterIP   10.107.154.9   <none>        443/TCP   32s

$ kubectl proxy &
$ open http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/#/login
```

While `proxy` works for some cases, it is often better to access services directly. Traditionally you could use [`minikube tunnel`](https://minikube.sigs.k8s.io/docs/handbook/accessing/), but that command only works from the same machine where the minikube Kubernetes is running. Instead, I ended up configuring a static route for the (*large*) minikube cluster IP range on my router (or locally with `route`).

{{< snippet dir="appendix/2020-04-11-minikube-and-bridged-networking" file="minikube.sh" lines="4" stripprefix="# " lang="bash" >}}

Once routed, I can directly access the dashboard by its Cluster IP.

```console
$ sudo tcptraceroute 10.107.154.9 443
Tracing the path to 10.107.154.9 on TCP port 443 (https), 30 hops max
 1  kubernetes.default.svc (192.0.2.113)  2.732 ms  1.108 ms  0.983 ms
 2  10.107.154.9  1.222 ms  1.075 ms  0.994 ms
 3  10.107.154.9 [open]  1.026 ms  1.514 ms  1.334 ms

$ open https://10.107.154.9/#/login
```

Success -- my minikube-managed Kubernetes cluster is now available for use by the rest of my local network!
