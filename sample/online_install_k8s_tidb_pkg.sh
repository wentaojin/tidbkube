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
Usage:
    $0 --master 192.168.10.4,192.168.10.5 --worker 192.168.10.6,192.168.10.7 --user root --password 123456
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

echo "check masterNodes: ${masterNodes[@]}"
echo "check workerNodes: ${workerNodes[@]}"
echo "check ssh user: ${sshUser}"
echo "check ssh password: ${sshPassword}"

# Minimum requirements soft version 
centosVersion="7.6"
kernelVersion="3.10.0"
dockerVersion="18.09.6"

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
dockerCmdList=(
"yum install -y yum-utils device-mapper-persistent-data lvm2"
"yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo"
"yum install docker-ce -y"
)

otherCmdList=(
"yum install ipset -y"
"yum install ipvsadm -y"
"yum install chrony -y"
"yum install -y kubelet kubeadm kubectl --disableexcludes=kubernetes"
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
       echo "host: ${host}, required greater or equal to docker verion: ${dockerVersion}, actual verion: ${dockerVersionRes}, would be reinstalling docker-ce"
       ${ssh_cmd} ${sshUser}@${host} ${dockerCmdList[0]}
       ${ssh_cmd} ${sshUser}@${host} ${dockerCmdList[1}
       ${ssh_cmd} ${sshUser}@${host} ${dockerCmdList[2]}         
      fi
    else
       ${ssh_cmd} ${sshUser}@${host} ${dockerCmdList[0]}
       ${ssh_cmd} ${sshUser}@${host} ${dockerCmdList[1}
       ${ssh_cmd} ${sshUser}@${host} ${dockerCmdList[2]}
    fi
done

# bootstrap the machine environment
for host in ${k8sNodes[@]} ;do
  echo "*****************************************************************"
  echo "          start host ${host} env bootstrap                       "
  echo "*****************************************************************"
  
  ${ssh_cmd} ${sshUser}@${host} "cat > /etc/yum.repos.d/kubernetes.repo" <<EOF 
[kubernetes]
name=Kubernetes
baseurl=http://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=0
repo_gpgcheck=0
gpgkey=http://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg
        http://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
EOF

  for cmd in "${otherCmdList}"; do
    ${ssh_cmd} ${sshUser}@${host} ${cmd}
  done
  
  ${ssh_cmd} ${sshUser}@${host} "systemctl start docker && systemctl enable docker && systemctl status docker"
  ${ssh_cmd} ${sshUser}@${host} "systemctl start chronyd && systemctl enable chronyd && systemctl status chronyd && chronyc sources"
  ${ssh_cmd} ${sshUser}@${host} "systemctl enable kubelet.service"
  
  echo -e "\033[32mbootstrap host ${host} env finished\033[0m"
done
