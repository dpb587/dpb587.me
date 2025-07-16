#!/bin/bash

# local routing
# sudo route add -net 10.96.0.0/12 $( get_bridged_ip )

dependencies() (
  echo brew
)

get_host_ip() (
  minikube ssh -- ip addr show eth1 | grep inet | awk '{ print $2 }' | cut -d/ -f1
)

get_bridged_ip() (
  minikube ssh -- ip addr show eth2 | grep inet | awk '{ print $2 }' | cut -d/ -f1
)

get_kubeconfig() (
  kubectl config view --flatten
)

applied() (
  _applied_install \
    && _applied_start \
    && _applied_bridged_nic \
    && _applied_hosts \
    || return 1
)

apply() (
  if ! _applied_install ; then
    DEBUG "installing minikube"
    _apply_install
  fi

  if ! _applied_start ; then
    DEBUG "starting"
    _apply_start
  fi

  if ! _applied_bridged_nic ; then
    DEBUG "configuring bridged nic"
    _apply_bridged_nic
  fi

  if ! _applied_hosts ; then
    DEBUG "updating hosts"
    _apply_hosts
  fi
)

_applied_install() (
  which -s minikube
)

_apply_install() (
  brew install minikube virtualbox
)

_applied_start() (
  tmpfile=/tmp/minikube.kubeconfig.$$
  get_kubeconfig > "$tmpfile"

  KUBECONFIG="$tmpfile" minikube update-context 2>&1 >/dev/null || true

  ret=0
  KUBECONFIG="$tmpfile" minikube status 2>&1 >/dev/null || ret="$?"

  rm "$tmpfile"

  return "$ret"
)

_apply_start() (
  minikube config set cpus "$( sysctl hw.ncpu | awk '{ print $2 }' )"
  minikube config set disk-size 65536
  minikube config set memory "$(( ( $( sysctl hw.memsize | awk '{ print $2 }' ) - 1073741824 ) / ( 1024 * 1024 ) ))"
  minikube config set vm-driver virtualbox

  minikube start
)

_applied_bridged_nic() (
  [ -n "$( VBoxManage showvminfo minikube --machinereadable 2>&1 | grep ^nic | grep 'bridged' | cut -d= -f1 | cut -c4- )" ]
)

_apply_bridged_nic() (
  VBoxManage controlvm minikube poweroff

  opts=""
  nic=$( VBoxManage showvminfo minikube --machinereadable | grep ^nic | grep '"none"' | head -n1 | cut -d= -f1 | cut -c4- )
  int=$( route get google.com | grep interface: | awk '{ print $2 }' )

  if [ -n "${MINIKUBE_NIC_BRIDGED_MAC:-}" ]; then
    opts="--macaddress$nic $MINIKUBE_NIC_BRIDGED_MAC"
  fi

  VBoxManage modifyvm minikube --nic$nic bridged --bridgeadapter$nic $int $opts

  minikube start
)

_applied_hosts() (
  desired="https://kubernetes.default.svc:8443"
  [[ "$( kubectl config view -o jsonpath='{.clusters[?(@.name == "minikube")].cluster.server}' )" == "$desired" ]] \
    || return 1

  desired="$( get_bridged_ip ) kubernetes.default.svc"
  grep -q "$desired" /etc/hosts \
    || return 1
)

_apply_hosts() (
  desired="https://kubernetes.default.svc:8443"
  kubectl config set-cluster minikube --server="$desired"

  desired="$( get_bridged_ip ) kubernetes.default.svc"
  sudo sed -i '' '/kubernetes.default.svc/d' /etc/hosts
  sudo tee -a /etc/hosts <<< "$desired" > /dev/null
)
