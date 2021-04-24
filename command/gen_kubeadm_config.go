package command

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/wentaojin/tidbkube/util"
)

const TemplateText = string(`apiVersion: kubeadm.k8s.io/v1beta1
kind: ClusterConfiguration
kubernetesVersion: {{.Version}}
controlPlaneEndpoint: "{{.ApiServer}}:6443"
imageRepository: {{.Repo}}
networking:
  # dnsDomain: cluster.local
  podSubnet: {{.PodCIDR}}
  serviceSubnet: {{.SvcCIDR}}
apiServer:
  certSANs:
  - 127.0.0.1
  - {{.ApiServer}}
  {{range .Masters -}}
  - {{.}}
  {{end -}}
  - {{.VIP}}
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
  - "{{.VIP}}/32"`)

var ConfigType string

func kubeadmConfig() string {
	var sb strings.Builder
	sb.Write([]byte(TemplateText))
	return sb.String()
}

func PrintlnKubeadmConfig(cobraFlag FlagCobra) {
	switch ConfigType {
	case "kubeadm":
		kubeadmConfigPrintln(cobraFlag)
	default:
		kubeadmConfigPrintln(cobraFlag)
	}
}

func kubeadmConfigPrintln(cobraFlag FlagCobra) {
	log.Println("Generate kubeadm config file,As follow:")
	fmt.Println(string(renderTemplateContent(kubeadmConfig(), cobraFlag)))
}

func Template(cobraFlag FlagCobra) []byte {
	return renderTemplateContent(kubeadmConfig(), cobraFlag)
}

func renderTemplateContent(templateContent string, cobraFlag FlagCobra) []byte {
	tmpl, err := template.New("text").Parse(templateContent)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Func renderTemplateContent:::Template parse failed: %v\n", err)
		}
	}()
	if err != nil {
		panic(err)
	}

	masters := util.ParseIPSegment(cobraFlag.MasterIP)
	var (
		envMap = make(map[string]interface{})
		buffer bytes.Buffer
	)
	envMap["VIP"] = cobraFlag.VirtualIP
	envMap["Masters"] = masters
	envMap["Version"] = cobraFlag.K8sVersion
	envMap["ApiServer"] = cobraFlag.ApiServer
	envMap["PodCIDR"] = cobraFlag.PodCIDR
	envMap["SvcCIDR"] = cobraFlag.SvcCIDR
	envMap["Repo"] = cobraFlag.ImageRepo

	err = tmpl.Execute(&buffer, envMap)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Func renderTemplateContent:::Template render failed: %v\n", err)
		}
	}()
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}
