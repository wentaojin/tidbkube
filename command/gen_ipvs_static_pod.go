package command

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/wentaojin/tidbkube/util"
)

const TemplateLVScare = string(`apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: null
  labels:
    component: kube-ipvs-lvscare
    tier: control-plane
  name: kube-ipvs-lvscare
  namespace: kube-system
spec:
  containers:
  - command:
    - /usr/bin/lvscare
    - care
    - --vs
    - {{.VIP}}:6443
    - --health-schem
    - https
    {{range .Masters -}}
    - --rs
    - {{.}}:6443
    {{end -}}
    image: fanux/lvscare:latest
    imagePullPolicy: IfNotPresent
    name: kube-ipvs-lvscare
    resources: {}
    securityContext:
      privileged: true
  hostNetwork: true
  priorityClassName: system-cluster-critical
status: {}`)

func ipvsStaticPodConfig() string {
	var sb strings.Builder
	sb.Write([]byte(TemplateLVScare))
	return sb.String()
}

func PrintlnIPVSStaticPodConfig(cobraFlag FlagCobra) {
	ipvsStaticPodConfigPrintln(cobraFlag)
}

func ipvsStaticPodConfigPrintln(cobraFlag FlagCobra) {
	log.Println("Generate ipvs static pod yaml config file,As follow:")
	fmt.Println(string(renderTemplateIPVSContent(ipvsStaticPodConfig(), cobraFlag)))
}

func TemplateIPVS(cobraFlag FlagCobra) []byte {
	return renderTemplateIPVSContent(ipvsStaticPodConfig(), cobraFlag)
}

func renderTemplateIPVSContent(templateContent string, cobraFlag FlagCobra) []byte {
	tmpl, err := template.New("text").Parse(templateContent)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Func renderTemplateIPVSContent:::Template parse failed: %v\n", err)
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

	err = tmpl.Execute(&buffer, envMap)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Func renderTemplateIPVSContent:::Template render failed: %v\n", err)
		}
	}()
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}
