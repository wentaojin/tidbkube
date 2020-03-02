package command

import (
	"fmt"
	"log"
	"strings"

	"github.com/WentaoJin/tidbkube/network"
)

const KubeadmDefaultTemplate = string(`apiVersion: kubeadm.k8s.io/v1beta1
kind: ClusterConfiguration
kubernetesVersion: v1.14.1
controlPlaneEndpoint: "apiserver.cluster.local:6443"
imageRepository: k8s.gcr.io
networking:
  # dnsDomain: cluster.local
  podSubnet: 100.64.0.0/10
  serviceSubnet: 10.96.0.0/12
apiServer:
  certSANs:
  - 127.0.0.1
  - apiserver.cluster.local
  - 10.103.97.4
  - 10.103.97.5
  - 10.103.97.6
  - 10.103.97.7
  - 10.103.97.8
  - 10.103.97.9
  - 10.103.97.2
  extraArgs:
    feature-gates: TTLAfterFinished=true
  extraVolumes:
  - name: localtime
    hostPath: /etc/localtime
    mountPath: /etc/localtime
    readOnly: true
    pathType: File
controllerManager:
  extraArgs:
    feature-gates: TTLAfterFinished=true
  extraVolumes:
  - hostPath: /etc/localtime
    mountPath: /etc/localtime
    name: localtime
    readOnly: true
    pathType: File
scheduler:
  extraArgs:
    feature-gates: TTLAfterFinished=true
  extraVolumes:
  - hostPath: /etc/localtime
    mountPath: /etc/localtime
    name: localtime
    readOnly: true
    pathType: File
---
apiVersion: kubeproxy.config.k8s.io/v1alpha1
kind: KubeProxyConfiguration
mode: "ipvs"
ipvs:
  excludeCIDRs: 
  - "10.103.97.2/32"`)

func PrintlnDefaultKubeadmTemplate() {
	log.Println("Print default kubeadm template config,As follow")
	fmt.Println()
	var sb strings.Builder
	sb.Write([]byte(KubeadmDefaultTemplate))
	fmt.Println(sb.String())
}

func PrintlnDefaultCalicoTemplate() {
	log.Println("Print default calico manifests template yaml config,As follow")
	fmt.Println()
	var sb strings.Builder
	sb.Write([]byte(network.CalicoManifests))
	fmt.Println(sb.String())
}

func PrintlnDefaultFlannelTemplate() {
	log.Println("Print default flannel manifests template yaml config,As follow")
	fmt.Println()
	var sb strings.Builder
	sb.Write([]byte(network.FlannelManifests))
	fmt.Println(sb.String())
}
