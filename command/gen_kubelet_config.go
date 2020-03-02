package command

import (
	"bytes"
	"html/template"
	"log"
	"strings"
)

const KubeletTemplate = string(`address: 0.0.0.0
apiVersion: kubelet.config.k8s.io/v1beta1
authentication:
  anonymous:
    enabled: false
  webhook:
    cacheTTL: 2m0s
    enabled: true
  x509:
    clientCAFile: /etc/kubernetes/pki/ca.crt
authorization:
  mode: Webhook
  webhook:
    cacheAuthorizedTTL: 5m0s
    cacheUnauthorizedTTL: 30s
cgroupDriver: {{.dockerDriver}}
cgroupsPerQOS: true
clusterDNS:
- 10.96.0.10
clusterDomain: cluster.local
configMapAndSecretChangeDetectionStrategy: Watch
containerLogMaxFiles: 5
containerLogMaxSize: 10Mi
contentType: application/vnd.kubernetes.protobuf
cpuCFSQuota: true
cpuCFSQuotaPeriod: 100ms
cpuManagerPolicy: none
cpuManagerReconcilePeriod: 10s
enableControllerAttachDetach: true
enableDebuggingHandlers: true
enforceNodeAllocatable:
- pods
eventBurst: 10
eventRecordQPS: 5
evictionHard:
  imagefs.available: 15%
  memory.available: 100Mi
  nodefs.available: 10%
  nodefs.inodesFree: 5%
evictionPressureTransitionPeriod: 5m0s
failSwapOn: true
fileCheckFrequency: 20s
hairpinMode: promiscuous-bridge
healthzBindAddress: 127.0.0.1
healthzPort: 10248
httpCheckFrequency: 20s
imageGCHighThresholdPercent: 85
imageGCLowThresholdPercent: 80
imageMinimumGCAge: 2m0s
iptablesDropBit: 15
iptablesMasqueradeBit: 14
kind: KubeletConfiguration
kubeAPIBurst: 10
kubeAPIQPS: 5
makeIPTablesUtilChains: true
maxOpenFiles: 1000000
maxPods: 110
nodeLeaseDurationSeconds: 40
nodeStatusUpdateFrequency: 10s
oomScoreAdj: -999
podPidsLimit: -1
port: 10255
registryBurst: 10
registryPullQPS: 5
resolvConf: /etc/resolv.conf
rotateCertificates: true
runtimeRequestTimeout: 2m0s
serializeImagePulls: true
staticPodPath: /etc/kubernetes/manifests
streamingConnectionIdleTimeout: 4h0m0s
syncFrequency: 1m0s
volumeStatsAggPeriod: 1m0s`)

func kubeletConfig() string {
	var sb strings.Builder
	sb.Write([]byte(KubeletTemplate))
	return sb.String()
}

func TemplateKubelet(cgroupDriver string) []byte {
	return renderTemplateKubeletContent(kubeletConfig(), cgroupDriver)
}

func renderTemplateKubeletContent(templateContent string, cgroupDriver string) []byte {
	tmpl, err := template.New("text").Parse(templateContent)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Func renderTemplateKubeletContent:::Template parse failed: %v\n", err)
		}
	}()
	if err != nil {
		panic(err)
	}

	var (
		envMap = make(map[string]interface{})
		buffer bytes.Buffer
	)
	envMap["dockerDriver"] = cgroupDriver

	err = tmpl.Execute(&buffer, envMap)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Func renderTemplateKubeletContent:::Template render failed: %v\n", err)
		}
	}()
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}
