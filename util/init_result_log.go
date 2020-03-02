package util

import (
	"strings"
)

// Host file check, output array with or without host file
func (sl *SSHResultLog) ResultCheckRemoteHostFileExist() (hostExist, hostNoExist []string) {
	hostRequired := checkHostResultValueFunc(sl, checkHostResultValueWorker)
	for _, resp := range hostRequired {
		if resp.Result.(string) == "0" {
			hostNoExist = append(hostNoExist, resp.Host)
		} else {
			hostExist = append(hostExist, resp.Host)
		}
	}
	return
}

func (sl *SSHResultLog) ResultCheckRemoteHostFileMd5(md5Sum string) (hostExist []string) {
	var (
		md5Slice []string
		md5Map   map[string]string
	)
	md5Map = make(map[string]string)
	hostRequired := checkHostResultValueFunc(sl, checkHostResultValueWorker)
	for _, resp := range hostRequired {
		md5Map[resp.Result.(string)] = resp.Host
		md5Slice = append(md5Slice, resp.Result.(string))
	}
	// determine whether they are the same file according to the md5sum value of the linux command
	for _, md5 := range md5Slice {
		if md5 != md5Sum {
			hostExist = append(hostExist, md5Map[md5])
		}
	}
	return hostExist
}

// decode master0 output to join token  hash and key
func (sl *SSHResultLog) ResultOutputKubeadmJoinParams() (joinToken, tokenCaCertHash string) {
	hostRequired := checkHostResultValueFunc(sl, checkHostResultValueWorker)
	joinToken, tokenCaCertHash = decodeJoinCmd(hostRequired[0].Result.(string))
	return
}

func (sl *SSHResultLog) ResultOutputKubeadmJoinMasterCert() (certificateKey string) {
	hostRequired := checkHostResultValueFunc(sl, checkHostResultValueWorker)
	certificateKey = decodeUploadCertsCmd(hostRequired[0].Result.(string))
	return
}

// 192.168.0.200:6443 --token 9vr73a.a8uxyaju799qwdjv
// --discovery-token-ca-cert-hash sha256:7c2e69131a36ae2a042a339b33381c6d0d43887e2de83720eff5359e26aec866 --control-plane
// --certificate-key f8902e114ef118304e561c3ecd4d0b543adc226b7a07f675f56564185ffe0c07
func decodeJoinCmd(joinCmd string) (joinToken, tokenCaCertHash string) {
	stringSlice := strings.Split(joinCmd, " ")
	for i, r := range stringSlice {
		switch r {
		case "--token":
			joinToken = strings.TrimSpace(stringSlice[i+1])
		case "--discovery-token-ca-cert-hash":
			tokenCaCertHash = strings.TrimSpace(stringSlice[i+1])
		}
	}
	return
}

func decodeUploadCertsCmd(joinCmd string) (certificateKey string) {
	joinSlice := strings.Split(joinCmd, "Using certificate key:\r\n")
	stringSlice := strings.Split(joinSlice[1], "\r\n")
	certificateKey = strings.TrimSpace(stringSlice[0])
	return
}

func (sl *SSHResultLog) ResultOutputCgroupDriver() (cgroupDriver string) {
	hostRequired := checkHostResultValueFunc(sl, checkHostResultValueWorker)
	cgroupDriver = hostRequired[0].Result.(string)
	return
}
