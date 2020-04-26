<center>Kubernentes TiDB 快速指南</center>

- master 节点允许调度（当前架构 1 master 2 worker,tidb cluster 要求 worker节点数至少 3 台）

  ```
  出于安全考虑，默认配置下Kubernetes不会将Pod调度到Master节点。如果希望将k8s-master也当作Node使用，可以执行如下命令
  
  $ kubectl taint node k8s-master node-role.kubernetes.io/master-
  
  其中 k8s-master 是主机节点 hostname 如果要恢复 Master Only状态，执行如下命令：
  
  $ kubectl taint node k8s-master node-role.kubernetes.io/master=""
  ```

- 部署 local pv （bind mount ）
  参考链接：https://github.com/kubernetes-sigs/sig-storage-local-static-provisioner/blob/master/docs/operations.md#sharing-a-disk-filesystem-by-multiple-filesystem-pvs

  ```
  DISK_UUID=$(blkid -s UUID -o value /dev/nvme1n1) 
  
  for i in $(seq 1 5); do
    sudo mkdir -p /data2/shared/${DISK_UUID}/vol${i} /mnt/shared/${DISK_UUID}_vol${i}
    sudo mount --bind /data2/shared/${DISK_UUID}/vol${i} /mnt/shared/${DISK_UUID}_vol${i}
  done
  
  for i in $(seq 1 5); do
    echo /data2/shared/${DISK_UUID}/vol${i} /mnt/shared/${DISK_UUID}_vol${i} none bind 0 0 | sudo tee -a /etc/fstab
  done
  ```

- RBAC (集群 1.16.6 默认开启，不用变动)

- 部署 helm

  ```
  $ helm version
    Client: &version.Version{SemVer:"v2.16.2", GitCommit:"bbdfe5e7803a12bbdf97e94cd847859890cf4050", GitTreeState:"clean"}
    Error: could not find an available port: listen tcp :0: bind: address already in use
  
  $ netstat -anlp | grep 44134
  tcp        0      0 172.16.5.89:44134       172.16.5.83:62479       ESTABLISHED 98179/bin/java (通信端口占用,重启该进程就会被释放)
  
  $ helm init --service-account=tiller --upgrade
  
  -- 再次查看
  $ helm version 
  Client: &version.Version{SemVer:"v2.16.2", GitCommit:"bbdfe5e7803a12bbdf97e94cd847859890cf4050", GitTreeState:"clean"}
  E0303 16:07:28.550177   69039 portforward.go:400] an error occurred forwarding 41716 -> 44134: error forwarding port 44134 to pod fd99c2cb51687acb0708503d5e8a4eacae6fecf8ceb7f885933872623f06cea5, uid : unable to do port forwarding: socat not found
  E0303 16:07:29.554231   69039 portforward.go:400] an error occurred forwarding 41716 -> 44134: error forwarding port 44134 to pod fd99c2cb51687acb0708503d5e8a4eacae6fecf8ceb7f885933872623f06cea5, uid : unable to do port forwarding: socat not found
  E0303 16:07:30.928658   69039 portforward.go:400] an error occurred forwarding 41716 -> 44134: error forwarding port 44134 to pod fd99c2cb51687acb0708503d5e8a4eacae6fecf8ceb7f885933872623f06cea5, uid : unable to do port forwarding: socat not found
  E0303 16:08:03.151281   69039 portforward.go:340] error creating error stream for port 41716 -> 44134: Timeout occured
  
  socat 没安装，所有 kubernentes 节点需要安装，yum install -y socat，问题解决
  ```

- helm 配置 https://pingcap.com/docs-cn/stable/tidb-in-kubernetes/reference/tools/in-kubernetes/#%E4%BD%BF%E7%94%A8-helm

  ```
  -- 添加 pingcap chasrts 仓库
  $ helm repo add pingcap https://charts.pingcap.org/
  ```

- 查询仓库应用依赖 Charts 版本 

  ```
  $ helm search pingcap -l
  NAME                    CHART VERSION   APP VERSION     DESCRIPTION                            
  pingcap/tidb-backup     latest                          A Helm chart for TiDB Backup or Restore
  pingcap/tidb-backup     v1.0.6                          A Helm chart for TiDB Backup or Restore
  pingcap/tidb-backup     v1.0.5                          A Helm chart for TiDB Backup or Restore
  pingcap/tidb-backup     v1.0.4                          A Helm chart for TiDB Backup or Restore
  pingcap/tidb-backup     v1.0.3                          A Helm chart for TiDB Backup or Restore
  pingcap/tidb-backup     v1.0.2                          A Helm chart for TiDB Backup or Restore
  pingcap/tidb-backup     v1.0.1                          A Helm chart for TiDB Backup or Restore
  pingcap/tidb-backup     v1.0.0                          A Helm chart for TiDB Backup or Restore
  pingcap/tidb-cluster    v1.0.6                          A Helm chart for TiDB Cluster          
  pingcap/tidb-cluster    v1.0.5                          A Helm chart for TiDB Cluster          
  pingcap/tidb-cluster    v1.0.4                          A Helm chart for TiDB Cluster          
  pingcap/tidb-cluster    v1.0.3                          A Helm chart for TiDB Cluster          
  pingcap/tidb-cluster    v1.0.2                          A Helm chart for TiDB Cluster          
  pingcap/tidb-cluster    v1.0.1                          A Helm chart for TiDB Cluster          
  pingcap/tidb-cluster    v1.0.0                          A Helm chart for TiDB Cluster          
  pingcap/tidb-cluster    latest                          A Helm chart for TiDB Cluster          
  pingcap/tidb-drainer    latest                          A Helm chart for TiDB Binlog drainer.  
  pingcap/tidb-drainer    dev                             A Helm chart for TiDB Binlog drainer.  
  pingcap/tidb-drainer    v1.0.6                          A Helm chart for TiDB Binlog drainer.  
  pingcap/tidb-drainer    v1.0.5                          A Helm chart for TiDB Binlog drainer.  
  pingcap/tidb-drainer    v1.0.4                          A Helm chart for TiDB Binlog drainer.  
  pingcap/tidb-lightning  dev                             A Helm chart for TiDB Lightning        
  pingcap/tidb-lightning  v1.0.6                          A Helm chart for TiDB Lightning        
  pingcap/tidb-lightning  v1.0.5                          A Helm chart for TiDB Lightning        
  pingcap/tidb-lightning  v1.0.4                          A Helm chart for TiDB Lightning        
  pingcap/tidb-lightning  latest                          A Helm chart for TiDB Lightning        
  pingcap/tidb-operator   v1.0.6                          tidb-operator Helm chart for Kubernetes
  pingcap/tidb-operator   v1.0.5                          tidb-operator Helm chart for Kubernetes
  pingcap/tidb-operator   v1.0.4                          tidb-operator Helm chart for Kubernetes
  pingcap/tidb-operator   v1.0.3                          tidb-operator Helm chart for Kubernetes
  pingcap/tidb-operator   latest                          tidb-operator Helm chart for Kubernetes
  pingcap/tidb-operator   v1.0.2                          tidb-operator Helm chart for Kubernetes
  pingcap/tidb-operator   v1.0.1                          tidb-operator Helm chart for Kubernetes
  pingcap/tidb-operator   v1.0.0                          tidb-operator Helm chart for Kubernetes
  ```

- 获取 tidb-operator yaml 文件

  ```
  $ helm inspect values pingcap/tidb-operator --version=v1.0.6 > values-tidb-operator.yaml
  ```

- tidb operato 镜像

  ```
  TiDB Operator 里面会用到 k8s.gcr.io/kube-scheduler 镜像，如果下载不了该镜像，可以修改 /home/tidb/tidb-operator/values-tidb-operator.yaml 文件中的 scheduler.kubeSchedulerImageName 为 registry.cn-hangzhou.aliyuncs.com/google_containers/kube-scheduler
  ```

- tidb operator 安装

  ```
  $ helm install pingcap/tidb-operator --name=tidb-operator --namespace=tidb-admin --version=v1.0.6 -f values-tidb-operator.yaml
  
  output:
  NOTES:
  
  1. Make sure tidb-operator components are running
     kubectl get pods --namespace tidb-admin -l app.kubernetes.io/instance=tidb-operator
  2. Install CRD
     kubectl apply -f https://raw.githubusercontent.com/pingcap/tidb-operator/master/manifests/crd.yaml
     kubectl get customresourcedefinitions
  3. Modify tidb-cluster/values.yaml and create a TiDB cluster by installing tidb-cluster charts
     helm install tidb-cluster
  ```

- 部署 tidb cluster

  ```
  -- 获取待安装的 tidb-cluster chart 的 values.yaml 配置文件以及配置
  $ helm inspect values pingcap/tidb-cluster --version=v1.0.6 > values-marvin-tidbv305.yaml
  
  values-<release-name>.yaml,release-name 将会作为 Kubernetes 相关资源（例如 Pod，Service 等）的前缀名，可以起一个方便记忆的名字，要求全局唯一，通过 helm ls -q 可以查看集群中已经有的 release-name
  
  -- 安装集群
  $ helm install pingcap/tidb-cluster --name=marvin-tidbv305 --namespace=marvin-tidbv305  --version=v1.0.6 -f values-marvin-tidbv305.yaml 
  
  Output:
  NOTES:
  Cluster Startup
  
  1. Watch tidb-cluster up and running
     watch kubectl get pods --namespace marvin-tidbv305 -l app.kubernetes.io/instance=marvin-tidbv305 -o wide
  2. List services in the tidb-cluster
     kubectl get services --namespace marvin-tidbv305 -l app.kubernetes.io/instance=marvin-tidbv305
  
  Cluster access
  
  - Access tidb-cluster using the MySQL client
    kubectl port-forward -n marvin-tidbv305 svc/marvin-tidbv305-tidb 4000:4000 &
    mysql -h 127.0.0.1 -P 4000 -u root -D test
    Set a password for your user
      SET PASSWORD FOR 'root'@'%' = '4Bm6FOOMS2'; FLUSH PRIVILEGES;
  - View monitor dashboard for TiDB cluster
    kubectl port-forward -n marvin-tidbv305 svc/marvin-tidbv305-grafana 3000:3000
    Open browser at http://localhost:3000. The default username and password is admin/admin.
    If you are running this from a remote machine, you must specify the server's external IP address.
  ```

- 查看 tidb cluster pod 状态

  ```
  $ kubectl get po -n marvin-tidbv305 -l app.kubernetes.io/instance=marvin-tidbv305
  NAME                                        READY   STATUS    RESTARTS   AGE
  marvin-tidbv305-discovery-df94699bf-52q9d   1/1     Running   0          62m
  marvin-tidbv305-monitor-6995cb488c-jb9gw    3/3     Running   0          62m
  marvin-tidbv305-pd-0                        1/1     Running   2          62m
  marvin-tidbv305-pd-1                        1/1     Running   1          62m
  marvin-tidbv305-pd-2                        1/1     Running   0          62m
  marvin-tidbv305-pump-0                      1/1     Running   1          62m
  marvin-tidbv305-pump-1                      1/1     Running   0          62m
  marvin-tidbv305-pump-2                      1/1     Running   0          61m
  marvin-tidbv305-tidb-0                      2/2     Running   0          59m
  marvin-tidbv305-tikv-0                      1/1     Running   0          60m
  marvin-tidbv305-tikv-1                      1/1     Running   0          60m
  marvin-tidbv305-tikv-2                      1/1     Running   0          60m
  ```

- 查看 NodePort 模式下对外暴露的IP/PORT（默认NodePort，externalTrafficPolicy: local 模式 只允许存在 tidb pod 才能访问）

  ```
  $ kubectl -n marvin-tidbv305 get svc marvin-tidbv305-tidb -ojsonpath="{.spec.ports[?(@.name=='mysql-client')].nodePort}{'\n'}"
  30404
  ```

- 变更所有主机可访问（externalTrafficPolicy: Cluster 模式）

  ```
  -- 编辑  values-marvin-tidbv305.yaml 文件
  tidb:
    service:
      type: NodePort
      externalTrafficPolicy: Cluster  # 添加该行
  
  $ helm upgrade marvin-tidbv305 pingcap/tidb-cluster --version=v1.0.6 -f values-marvin-tidbv305.yaml 
  ```

- 查看 tidb 监控端口

  ```
  -- 默认 NodePort Cluster 模式，任何主机 + 端口即可访问监控
  $ kubectl -n marvin-tidbv305 get svc marvin-tidbv305-grafana
  NAME                      TYPE       CLUSTER-IP     EXTERNAL-IP   PORT(S)          AGE
  marvin-tidbv305-grafana   NodePort   10.106.59.66   <none>        3000:31431/TCP   47m
  ```

- 查看 prometheus 端口

  ```
  -- 默认 NodePort Cluster 模式，任何主机 + 端口即可访问 Prometheus
  $ kubectl -n marvin-tidbv305 get svc marvin-tidbv305-prometheus
  NAME                         TYPE       CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
  marvin-tidbv305-prometheus   NodePort   10.102.168.177   <none>        9090:30977/TCP   48m
  ```

  