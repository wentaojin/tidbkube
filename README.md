# tidbkube
tidbkube is used for deploy k8s with kubeadm and bootstrap tidb environment.

### 使用说明

- 全局命令

```
Welcome to
 _____ _ ____  ____  _  __     _
|_   _(_)  _ \| __ )| |/ /   _| |__   ___
  | | | | | | |  _ \| ' / | | | '_ \ / _ \
  | | | | |_| | |_) | . \ |_| | |_) |  __/
  |_| |_|____/|____/|_|\_\__,_|_.__/ \___|

Program tidbkube is an application to quickly set up a kubernentes cluster and initialize the
kubernentes cluster machine environment according to the tidb database requirements

Usage:
  tidbkube [command]

Available Commands:
  bootstrap   初始化主机集群环境
  execute     手工运行批量运行 shell 命令
  help        Help about any command
  inspect     检查主机系统环境是否符合要求或者部分软件是否提前安装
  kube        用于 kubernentes 集群得安装、节点 Join、集群重置 Reset

Flags:
  -h, --help                 help for tidbkube
      --list-task            列举命令运行得所有任务名
  -m, --master strings       指定 kubernentes 集群 master 节点，多个 master 形式 -m 192.168.0.2 -m 192.168.0.3 -m 192.168.0.3
  -p, --password string      指定 kubernentes 集群主机间得 SSH root 用户密码
  -P, --port string          指定 kubernentes 集群主机间得 SSH 端口，默认 22
  -k, --private-key string   指定 kubernentes 集群主机间得 private key,默认 /root/.ssh/id_rsa
      --skip-task string     指定命令跳过指定任务运行 (例如: bootstrap_system_package,bootstrap_chrony_server)
      --task string          指定命令只运行指定任务 (for example: check_system_cpu,check_system_version)
  -u, --user string          指定 kubernentes 集群主机间得 SSH root 用户
  -w, --worker strings        指定 kubernentes 集群 worker 节点，多个 worker 形式 -w 192.168.0.4 -w 192.168.0.5 -w 192.168.0.6

Use "tidbkube [command] --help" for more information about a command.
```

- 集群 kube 命令

```
Welcome to
 _____ _ ____  ____  _  __     _
|_   _(_)  _ \| __ )| |/ /   _| |__   ___
  | | | | | | |  _ \| ' / | | | '_ \ / _ \
  | | | | |_| | |_) | . \ |_| | |_) |  __/
  |_| |_|____/|____/|_|\_\__,_|_.__/ \___|

The command kube is used for kubernentes init to initialize the cluster installation and join, and add kubernentes nodes

Usage:
  tidbkube kube [command]

Aliases:
  kube, kb

Available Commands:
  init        kubernentes 集群安装
  join        kubernentes 集群节点加入
  reset       kubernentes 集群重置,不包括系统环境得重置(bootstrap)

Flags:
      --apiserver string           指定控制节点 Master DNS 名字，一般不用指定，保持默认 (default "apiserver.cluster.local")
      --control-plane              Kubernetes 集群执行 join 命令加入 master 节点,需要设置，否则默认以 worker 节点加入集群
  -h, --help                       help for kube
      --join-node strings          Kubernetes 集群执行 join 命令加入 master 节点，多个 master 加入 --join-node 192.168.0.2 --join-worker 192.168.0.5
  -v, --k8s-version string         Kubernentes 集群指定安装版本，形式 -v v1.16.6（ flag 必须指定）
  -c, --kubeadm-config string      指定 kubeadm 安装配置文件 Kubeadm-config.yaml，一般不用设置，自动会生成
      --net-interface string       指定主机 IP 地址网络接口名称，默认(default "eth.*|em.*")，若接口名不是这两者其中一个，则需要指定
      --net-plugin-config string   指定 Kubernentes network plugin 配置文件，默认内部安装 calico 插件，无需特别要求，不建议指定(for example: /root/kube-flannel.yaml)
  -n, --net-plugin-name string     指定 Kubernentes network plugin 名(default "calico")
      --pkg-path string            指定离线安装包路径 (default "/root/kube1.14.1.tar.gz")
      --podcidr string             k8s集群中pod的ip地址范围，一般不用更改 (default "100.64.0.0/10")
      --repo string                集群相关容器镜像仓库地址，从该地址拉取容器，默认镜像 docker load 加载本地，若无特殊要求可不设置 (default "k8s.gcr.io")
      --svccidr string             k8s集群中service地址范围，一般不用更改 (default "10.96.0.0/12")
      --vip string                 kubernentes 虚拟 IP，一般不用更改(default "10.103.97.2")
      --without-cni                如果不安装 calico 网络插件，则需要设置（默认 false）,另外还需设置--net-plugin-config、--net-plugin-name
```

- 示例安装

```
-- 环境检查
./tidbkube inspect -m 192.168.0.2 -m 192.168.0.3 -m 192.168.0.4 -w 192.168.0.5 -w 192.168.0.6 -w 192.168.0.7 -u root -p 123456 -P 22

-- 环境初始化
./tidbkube bootstrap -m 192.168.0.2 -m 192.168.0.3 -m 192.168.0.4 -w 192.168.0.5 -w 192.168.0.6 -w 192.168.0.7 -u root -p 123456 -P 22

-- 集群安装
./tidbkube kube init -m 192.168.0.2 -m 192.168.0.3 -m 192.168.0.4 -w 192.168.0.5 -w 192.168.0.6 -w 192.168.0.7 -u root -p 123456 -P 22 -v v1.16.6 --pkg-path /data/marvin/kube1.16.6.tar.gz

-- 集群节点 Join(master join 与 worker join 不能同时进行)
Node Join 需要注意把所有存在 master、worker 写上，需要配置 ipvs 规则以及 static pod yaml 文件
- join master：
./tidbkube kube join -m 192.168.0.2 -m 192.168.0.3 -m 192.168.0.4 -w 192.168.0.5 -w 192.168.0.6 -w 192.168.0.7 --join-node 192.168.0.8 -u root -p 123456 -P 22 -v v1.16.6 --pkg-path /data/marvin/kube1.16.6.tar.gz --control-plane

- join worker：
./tidbkube kube join -m 192.168.0.2 -m 192.168.0.3 -m 192.168.0.4 -w 192.168.0.5 -w 192.168.0.6 -w 192.168.0.7 --join-node 192.168.0.9 -u root -p 123456 -P 22 -v v1.16.6 --pkg-path /data/marvin/kube1.16.6.tar.gz

-- 清理集群
Reset 集群需要把所有存在 master、worker 写上
./tidbkube kube reset  -m 192.168.0.2 -m 192.168.0.3 -m 192.168.0.4 -w 192.168.0.5 -w 192.168.0.6 -w 192.168.0.7 -u root -p 123456 -P 22

可选项： 
--remove                   清理集群，包括安装包、docker imasge、kubernentes binary，默认该三者不清理
--remove-container-images  清理集群，只包括 docker imasge
--remove-install-pkg       清理集群，只包括安装包
--remove-kube-component    清理集群，只包括kubernenets binary
```

- 结果输出（单主多worker）

```
->kubectl get pod --all-namespaces -n kube-system
NAMESPACE     NAME                                       READY   STATUS    RESTARTS   AGE
kube-system   calico-kube-controllers-688c5dc8c7-7nq6l   1/1     Running   0          15h
kube-system   calico-node-bn5lb                          1/1     Running   0          15h
kube-system   calico-node-gz9z4                          1/1     Running   0          88m
kube-system   calico-node-qq47b                          1/1     Running   0          15h
kube-system   calico-node-wss86                          1/1     Running   0          15h
kube-system   coredns-5644d7b6d9-thh4w                   1/1     Running   0          15h
kube-system   coredns-5644d7b6d9-vdzfr                   1/1     Running   0          15h
kube-system   etcd-test3                                 1/1     Running   0          15h
kube-system   kube-apiserver-test3                       1/1     Running   0          15h
kube-system   kube-controller-manager-test3              1/1     Running   0          15h
kube-system   kube-ipvs-lvscare-test1                    1/1     Running   0          15h
kube-system   kube-ipvs-lvscare-test2                    1/1     Running   0          15h
kube-system   kube-ipvs-lvscare-test4                    1/1     Running   0          15h
kube-system   kube-proxy-6qzgt                           1/1     Running   0          92m
kube-system   kube-proxy-7krfb                           1/1     Running   0          15h
kube-system   kube-proxy-9cwr5                           1/1     Running   0          15h
kube-system   kube-proxy-hz6rm                           1/1     Running   0          15h
kube-system   kube-scheduler-test3                       1/1     Running   0          15h
```

