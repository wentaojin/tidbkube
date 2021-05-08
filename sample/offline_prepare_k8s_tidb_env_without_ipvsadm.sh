#!/usr/bin/bash

# Copyright 2020 PingCAP, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# See the License for the specific language governing permissions and
# limitations under the License.

set -e
set +o errexit

usage() {
    cat <<EOF
This script use to check machine environment before kubernetes cluster deploy
Before run this script,please ensure that:
* have installed sshpass
* machines ssh user password need to be unified
Options:
    -h,--help              prints the usage message
    -m,--master            master nodes of the kubernetes cluster,sample: 192.168.10.4,192.168.10.5
    -w,--worker            worker nodes of the kubernetes cluster,sample: 192.168.10.6,192.168.10.7
    -u,--user              ssh user,default: root
    -p,--password          ssh user password,sample: 123456
    -f,--firewalld         whether the firewall can be turned off,default: off
    -d,--data-root         docker deaemon update, configure images and container log store dir,default: /home
Usage:
    $0 --master 192.168.10.4,192.168.10.5 --worker 192.168.10.6,192.168.10.7 --user root --password 123456 --firewalld off --data-root /home
EOF
}

if [ $# -eq 0 ]; then
  usage
  exit 1
fi

while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    -m|--master)
    masterNodes="$2"
    shift
    shift
    ;;
    -w|--worker)
    workerNodes="$2"
    shift
    shift
    ;;
    -u|--user)
    sshUser="$2"
    shift
    shift
    ;;
    -p|--password)
    sshPassword="$2"
    shift
    shift
    ;;
    -f|--firewalld)
    firewalld="$2"
    shift
    shift
    ;;
    -d|--data-root)
    dataRoot="$2"
    shift
    shift
    ;;
    -h|--help)
    usage
    exit 0
    ;;
    *)
    echo "unknown option: $key"
    usage
    exit 1
    ;;
esac
done

# Kubernents deploy machines
#masterNodes=(${masterNodes//,/})
#workerNodes=(${workerNodes//,/})
masterNodes=(`echo ${masterNodes} | tr ',' ' '`)
workerNodes=(`echo ${workerNodes} | tr ',' ' '`)
sshUser=${sshUser:-root}
sshPassword=${sshPassword}
firewalld=${firewalld:-off}
dataRoot=${dataRoot:-/home}

# Minimum requirements soft version 
centosVersion="7.6"
kernelVersion="3.10.0"
dockerVersion="18.09.6"

echo "prepare masterNodes: ${masterNodes[@]}"
echo "prepare workerNodes: ${workerNodes[@]}"
echo "prepare ssh user: ${sshUser}"
echo "prepare ssh password: ${sshPassword}"
echo "firewalld whether is turnning off: ${firewalld}"
echo "docker data root: ${dataRoot}"


# Package sshpass check installed
function err_exit() {
  echo "ERROR: $1" 1>&2
  exit 1
}
which sshpass >/dev/null 2>&1 || err_exit "current run script machine sshpass not installed"

# Function
# Version compare
# V1 > V2
function version_gt() { test "$(echo "$@" | tr " " "\n" | sort -V | head -n 1)" != "$1"; }
# V1 >= V2
function version_ge() { test "$(echo "$@" | tr " " "\n" | sort -rV | head -n 1)" == "$1"; }
# V1 <= V2
function version_le() { test "$(echo "$@" | tr " " "\n" | sort -V | head -n 1)" == "$1"; }
# V1 < V2
function version_lt() { test "$(echo "$@" | tr " " "\n" | sort -rV | head -n 1)" != "$1"; } 

# Kubernents machine nodes
k8sNode=(${masterNodes[@]}  ${workerNodes[@]})
# Remove repeat host machine list
k8sNodes=($(echo ${k8sNode[*]} | sed 's/ /\n/g' | sort | uniq))

# Set need exec shell cmd
firewalldCmdList=(
"systemctl stop firewalld && systemctl disable firewalld && systemctl status firewalld"
"systemctl start firewalld && systemctl enable firewalld && systemctl status firewalld"
)
otherCmdList=(
"setenforce 0 && sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config"
"swapoff -a && sed -i 's/^\(.*swap.*\)$/#\1/' /etc/fstab"
"systemctl enable irqbalance && systemctl start irqbalance && systemctl status irqbalance"
"cpupower frequency-set --governor performance"
"systemctl enable docker && sed -i 's/^LimitNOFILE=infinity/LimitNOFILE=1048576/g' /usr/lib/systemd/system/docker.service && systemctl daemon-reload && systemctl restart docker && grep 'LimitNOFILE=1048576' /usr/lib/systemd/system/docker.service"
)
firewalldOpenCmdList=(
"firewall-cmd --permanent --add-port=6443/tcp"
"firewall-cmd --permanent --add-port=2379-2380/tcp"
"firewall-cmd --permanent --add-port=10250/tcp"
"firewall-cmd --permanent --add-port=10251/tcp"
"firewall-cmd --permanent --add-port=10252/tcp"
"firewall-cmd --permanent --add-port=10255/tcp"
"firewall-cmd --permanent --add-port=8472/udp"
"firewall-cmd --add-masquerade --permanent"
"firewall-cmd --permanent --add-port=30000-32767/tcp"
"systemctl restart firewalld"
)

#Configure ssh params 
#Ordinary ssh option, the same task execution time is basically the same each time (used in OpenSSH version before 5.6)
#ssh_options=" -o StrictHostKeyChecking=no"
ssh_options="  -o StrictHostKeyChecking=no -o PubkeyAuthentication=no "
#Flash ssh optimization option, using ssh long connection multiplexing technology, 
#the same task is within 10s, and the second and subsequent execution time is within 3s (used in OpenSSH 5.6 and later)
#ssh_options=" -T -q -o StrictHostKeyChecking=no -o PubkeyAuthentication=no -o ConnectTimeout=5  -o ControlMaster=auto -o ControlPath=$tmp_dir/.ssh_mux_%h_%p_%r -o ControlPersist=600s "

#Flash ssh optimization option, enable compression option, use in low-speed network link environment, 
#use ssh long connection multiplexing technology, the same task within 10s, the second and subsequent execution time within 3s (used in OpenSSH 5.6 and later)
#ssh_options=" -C -tt -q -o StrictHostKeyChecking=no -o PubkeyAuthentication=no -o ConnectTimeout=5  -o ControlMaster=auto -o ControlPath=$tmp_dir/.ssh_mux_%h_%p_%r -o ControlPersist=600s"
ssh_cmd="/usr/bin/sshpass -p${sshPassword} /usr/bin/ssh ${ssh_options}"


# Check soft version requirements
for host in ${k8sNodes[@]} ;do
    centosVersionRes=`${ssh_cmd} ${sshUser}@${host} "cat /etc/redhat-release" | awk '{print $4}'`
    kernelVersionRes=`${ssh_cmd} ${sshUser}@${host} "uname -r" | awk -F '-' '{print $1}'`
    dockerVersionRes=`${ssh_cmd} ${sshUser}@${host} "docker -v" | awk '{print $3}'| awk -F ',' '{print $1}' 2>/dev/null || echo '0'`
    if version_lt ${centosVersionRes} ${centosVersion}; then
      echo "host: ${host}, required greater or equal to os verion: ${centosVersion}, actual verion: ${centosVersionRes}, need upgrade os version or replace os"
      exit 1
    fi
    if version_lt ${kernelVersionRes} ${kernelVersion}; then
      echo "host: ${host}, required greater or equal to os kernel verion: ${kernelVersion}, actual verion: ${kernelVersionRes}, need upgrade os kernel or replace os"
      exit 1
    fi
    if [ "${dockerVersionRes}" != "0" ]; then
      if version_lt ${dockerVersionRes} ${dockerVersion}; then
       echo "host: ${host}, required greater or equal to docker verion: ${dockerVersion}, actual verion: ${dockerVersionRes}, need be uinstall docker-ce"
       exit 1    
      fi
    fi
done

# Prepare /etc/hosts file
declare -a hostNames
declare -A map=()

for host in ${k8sNodes[@]} ;do
  name=`${ssh_cmd} ${sshUser}@${host} hostname`
  map["${host}"]="${name}"
  hostNames=(${hostNames[*]} ${host})
done


for host in ${k8sNodes[@]} ;do
  echo "*****************************************************************"
  echo "          start host ${host} hosts file prepare                  "
  echo "*****************************************************************"   
  for name in "${hostNames[@]}" ;do
    checkIsExist=`${ssh_cmd} ${sshUser}@${host} grep -E \"${name} *${map[${name}]}\" /etc/hosts | sort -u | uniq | wc -l`
    if [ "${checkIsExist}" != "1" ]; then
      ${ssh_cmd} ${sshUser}@${host} "cat >> /etc/hosts" << EOF
${name} ${map[${name}]}
EOF
    fi  
 done
 echo -e "\033[32mprepare host ${host} hosts file success\033[0m"
done

# Initialize the environment
for host in ${k8sNodes[@]} ;do
  echo "*****************************************************************"
  echo "          start host ${host} env prepare                         "
  echo "*****************************************************************"
  echo "prepare firewalld"
  if [ "${firewalld}" = "off" ]; then
    ${ssh_cmd} ${sshUser}@${host} ${firewalldCmdList[0]}
  else
    ${ssh_cmd} ${sshUser}@${host} ${firewalldCmdList[1]}

    for cmd in "${firewalldOpenCmdList[@]}" ;do
      ${ssh_cmd} ${sshUser}@${host} ${cmd}
    done

  fi

  echo -e "\033[32m*\033[0m"
  echo "prepare security limits.conf"
  limitSoftNofileConf=`${ssh_cmd} ${sshUser}@${host} "cat  /etc/security/limits.conf | grep -E 'root *soft *nofile'| sort -u | uniq | wc -l"`
  limitHardNofileConf=`${ssh_cmd} ${sshUser}@${host} "cat  /etc/security/limits.conf | grep -E 'root *hard *nofile'| sort -u | uniq | wc -l"`
  limitSoftStackConf=`${ssh_cmd} ${sshUser}@${host} "cat  /etc/security/limits.conf | grep -E 'root *soft *stack'| sort -u | uniq | wc -l"`
  if [ "${limitSoftNofileConf}" != "1" ]; then
    ${ssh_cmd} ${sshUser}@${host} "echo 'root        soft        nofile        1048576' >> /etc/security/limits.conf"
  fi
  if [ "${limitHardNofileConf}" != "1" ]; then
    ${ssh_cmd} ${sshUser}@${host} "echo 'root        hard        nofile        1048576' >> /etc/security/limits.conf"
  fi
  if [ "${limitSoftStackConf}" != "1" ]; then
    ${ssh_cmd} ${sshUser}@${host} "echo 'root        soft        stack         10240' >> /etc/security/limits.conf"
  fi
  ${ssh_cmd} ${sshUser}@${host} "grep 'root' /etc/security/limits.conf"

  echo -e "\033[32m*\033[0m"
  echo "prepare iptables"
  iptablesConf=`${ssh_cmd} ${sshUser}@${host} "grep 'iptables -P FORWARD ACCEPT' /etc/rc.local | sort -u | uniq | wc -l"`
  if [ "${iptablesConf}" != "1" ]; then
    ${ssh_cmd} ${sshUser}@${host} "iptables -P FORWARD ACCEPT && chmod +x /etc/rc.local && echo 'iptables -P FORWARD ACCEPT' >> /etc/rc.local"
  fi
  
  echo -e "\033[32m*\033[0m"
  echo "prepare selinux swap irqbalance docker"
  for cmd in "${otherCmdList[@]}"; do
    ${ssh_cmd} ${sshUser}@${host} ${cmd}
  done
  
  echo -e "\033[32m*\033[0m"
  echo "prepare sysctl conf"
  ${ssh_cmd} ${sshUser}@${host} "cat > /etc/sysctl.d/k8s.conf" << EOF
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
vm.swappiness = 0
net.ipv4.tcp_syncookies = 0
net.ipv4.ip_forward = 1
fs.file-max = 1000000
fs.inotify.max_user_watches = 1048576
fs.inotify.max_user_instances = 1024
net.ipv4.conf.all.rp_filter = 1
net.ipv4.neigh.default.gc_thresh1 = 80000
net.ipv4.neigh.default.gc_thresh2 = 90000
net.ipv4.neigh.default.gc_thresh3 = 100000
EOF

  echo -e "\033[32m*\033[0m"
  echo "prepare docker daemon.json"
  ${ssh_cmd} ${sshUser}@${host} "cat > /etc/docker/daemon.json" << EOF
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2",
  "storage-opts": [
    "overlay2.override_kernel_check=true"
  ],
  "data-root": "${dataRoot}"
}
EOF
  
  ${ssh_cmd} ${sshUser}@${host} "cat /etc/docker/daemon.json"
  ${ssh_cmd} ${sshUser}@${host} "modprobe br_netfilter"
  ${ssh_cmd} ${sshUser}@${host} "sysctl -p /etc/sysctl.d/k8s.conf"
  ${ssh_cmd} ${sshUser}@${host} "sysctl --system"
  
  echo -e "\033[32m*\033[0m"
  echo -e "\033[32mprepare host ${host} env success\033[0m"
done
