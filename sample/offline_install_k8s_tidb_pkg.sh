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
    -d,--dockerPkg         docker-ce offline pkg tar path
    -k,--k8sPkg            k8s component offline pkg tar path
    -t,--tmp               component untar tmp store path
Usage:
    $0 --master 192.168.10.4,192.168.10.5 --worker 192.168.10.6,192.168.10.7 --user root --password 123456 --dockerPkg /home/docker-19.4.tar.gz --k8sPkg /home/k8s.tar.gz --tmp /tmp
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
    -d|--dockerPkg)
    dockerPkg="$2"
    shift
    shift
    ;;
    -k|--k8sPkg)
    k8sPkg="$2"
    shift
    shift
    ;;
    -t|--tmp)
    tmpDir="$2"
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
dockerPkg=${dockerPkg:-/home/docker-19.4.tar.gz}
k8sPkg=${k8sPkg:-/home/k8s.tar.gz}
tmpDir=${tmp:-/tmp}

echo "check masterNodes: ${masterNodes[@]}"
echo "check workerNodes: ${workerNodes[@]}"
echo "check ssh user: ${sshUser}"
echo "check ssh password: ${sshPassword}"
echo "docker pkg path: ${dockerPkg}"
echo "k8s pkg path: ${k8sPkg}"
echo "pkg untar tmp store path: ${tmpDir}"


# component pkg untar
tar -zxvf ${k8sPkg} -C ${tmpDir}
tar -zxvf ${dockerPkg} -C ${tmpDir}

scpDockerDir=${tmpDir}/docker/*
scpK8sDir=${tmpDir}/kubernetes/node/bin/

# Kubernents machine nodes
k8sNode=(${masterNodes[@]}  ${workerNodes[@]})
# Remove repeat host machine list
k8sNodes=($(echo ${k8sNode[*]} | sed 's/ /\n/g' | sort | uniq))

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
scp_cmd="/usr/bin/sshpass -p${sshPassword} /usr/bin/scp -r "

# docker component install
for host in ${k8sNodes[@]} ;do
  echo "*****************************************************************"
  echo "          start host ${host} docker offline install              "
  echo "*****************************************************************"
  ${scp_cmd} ${scpDockerDir}  ${sshUser}@${host}:/usr/bin
  
  ${ssh_cmd} ${sshUser}@${host} "cat > /etc/systemd/system/docker.service" <<EOF 
[Unit]
Description=Docker Application Container Engine
Documentation=https://docs.docker.com
After=network-online.target firewalld.service
Wants=network-online.target

[Service]
Type=notify
# the default is not to use systemd for cgroups because the delegate issues still
# exists and systemd currently does not support the cgroup feature set required
# for containers run by docker
ExecStart=/usr/bin/dockerd
ExecReload=/bin/kill -s HUP $MAINPID
# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity
# Uncomment TasksMax if your systemd version supports it.
# Only systemd 226 and above support this version.
#TasksMax=infinity
TimeoutStartSec=0
# set delegate yes so that systemd does not reset the cgroups of docker containers
Delegate=yes
# kill only the docker process, not all processes in the cgroup
KillMode=process
# restart the docker process if it exits prematurely
Restart=on-failure
StartLimitBurst=3
StartLimitInterval=60s

[Install]
WantedBy=multi-user.target
EOF

  ${ssh_cmd} ${sshUser}@${host} "chmod +x /etc/systemd/system/docker.service"

  ${ssh_cmd} ${sshUser}@${host} "systemctl daemon-reload && systemctl start docker && systemctl enable docker.service"

  ${ssh_cmd} ${sshUser}@${host} "docker -v"
  
   echo -e "\033[32mhost ${host} docker install finished\033[0m"
done

# k8s component
for host in ${masterNodes[@]} ;do
  echo "*****************************************************************"
  echo "          start host ${host} k8s pkg offline install             "
  echo "*****************************************************************"
  ${scp_cmd} ${scpK8sDir}/kubeadm  ${sshUser}@${host}:/usr/bin
  ${scp_cmd} ${scpK8sDir}/kubectl  ${sshUser}@${host}:/usr/bin  
  ${scp_cmd} ${scpK8sDir}/kubelet  ${sshUser}@${host}:/usr/bin

  ${ssh_cmd} ${sshUser}@${host} "kubeadm version"
  echo -e "\033[32mhost ${host} master pkg install finished\033[0m"
done

for host in ${workerNodes[@]} ;do
  echo "*****************************************************************"
  echo "          start host ${host} k8s pkg offline install             "
  echo "*****************************************************************"
  ${scp_cmd} ${scpK8sDir}/kubeadm  ${sshUser}@${host}:/usr/bin
  ${scp_cmd} ${scpK8sDir}/kubectl  ${sshUser}@${host}:/usr/bin  
  ${scp_cmd} ${scpK8sDir}/kubelet  ${sshUser}@${host}:/usr/bin

  ${ssh_cmd} ${sshUser}@${host} "kubelet --version"
  echo -e "\033[32mhost ${host} worker pkg install finished\033[0m"
done
