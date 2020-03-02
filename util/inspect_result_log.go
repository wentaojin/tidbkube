package util

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
)

// HostRequired struct
type HostRequired struct {
	Host   string
	Result interface{}
}

//This go file is mainly used to check the host environment requirements for the successful hosts in the SSH result log.
// If the requirements are met, the remaining commands continue to be executed. If the requirements are not met, the process is terminated.
// It is also necessary to pay attention to all functions in this file The execution is only performed on the successful host,
// the error host is not checked, because the error host check has been detected before the go file, no need to check
func resultCheckSystemVersion(sl *SSHResultLog) {
	var hostNoRequire []HostRequired
	hostRequired := checkHostResultValueFunc(sl, checkSystemVersionWorker)
	for _, resp := range hostRequired {
		if resp.Result.(float64) < 7.4 || resp.Result.(float64) > 8.0 {
			hostNoRequire = append(hostNoRequire, resp)
		}
	}
	checkHostResultOutput(hostNoRequire, "Func resultCheckHostOSVersion:::Host %s OS version %v non-compliant,"+
		"Require higher than or equal to 7.4 and lower 8", "")
}

func resultCheckSystemCPU(sl *SSHResultLog, requiredIntValue int) {
	var hostNoRequire []HostRequired
	requiredValue := strconv.Itoa(requiredIntValue)
	hostRequired := checkHostResultValueFunc(sl, checkHostResultValueWorker)
	for _, resp := range hostRequired {
		if resp.Result.(string) < requiredValue {
			hostNoRequire = append(hostNoRequire, resp)
		}
	}
	checkHostResultOutput(hostNoRequire, "Func resultCheckHostCPUSize:::Host %s OS logic cpu %v non-compliant,"+
		"Require higher than or equal to %dVC", requiredValue)

}

func resultCheckSystemMemory(sl *SSHResultLog, requiredIntValue int) {
	var hostNoRequire []HostRequired
	requiredValue := strconv.Itoa(requiredIntValue)
	hostRequired := checkHostResultValueFunc(sl, checkHostResultValueWorker)
	for _, resp := range hostRequired {
		if resp.Result.(string) < requiredValue {
			hostNoRequire = append(hostNoRequire, resp)
		}
	}
	checkHostResultOutput(hostNoRequire, "Func resultCheckHostOSMemory:::Host %s OS memory size %v non-compliant,"+
		"Require higher than or equal to %dGB", requiredValue)
}

func resultCheckSystemNetwork(sl *SSHResultLog, requiredStringValue string) {
	var noRequireRes []string

	hostRequired := checkHostResultValueFunc(sl, checkSystemNetworkWorker)
	for _, resp := range hostRequired {
		for _, v := range resp.Result.([]string) {
			// regular match ping result string percentage, ie packet loss in ping output,
			// if value is equal to 0%,show network normal,not exist packet loss
			if v != requiredStringValue {
				noRequireRes = append(noRequireRes, v)
			}
		}
	}
	if len(noRequireRes) != 0 {
		log.Printf("Func resultCheckSystemNetwork:::There is packet loss between the host and other hosts, "+
			"which does not meet the requirements,Packet loss value slice %v,Require packet loss is equal to %s",
			noRequireRes, requiredStringValue)
		os.Exit(1)
	}
}

func resultCheckSystemHostname(sl *SSHResultLog) {
	var (
		hostRes []string
	)
	hostRequired := checkHostResultValueFunc(sl, checkHostResultValueWorker)
	for _, resp := range hostRequired {
		switch rt := resp.Result.(type) {
		case string:
			hostRes = append(hostRes, rt)
		case []string:
			hostRes = append(hostRes, rt...)
		default:
			log.Println("Func resultCheckSystemHostname:::Result type check Unknown")
		}
	}

	hostResRemoveRepeat := StringSliceRemoveRepeat(hostRes)

	hostResRepeat := DiffStringSlices(hostResRemoveRepeat, hostRes)
	if len(hostResRepeat) != 0 {
		log.Printf("Func resultCheckSystemHostname:::Host system hostname exist repeat value %v, "+
			"Require hostname non-repeat", hostResRepeat)
		os.Exit(1)
	}
}

func resultCheckSystemNetworkCardMAC(sl *SSHResultLog) {
	var (
		hostRes []string
	)
	hostRequired := checkHostResultValueFunc(sl, checkHostResultValueWorker)
	for _, resp := range hostRequired {
		switch rt := resp.Result.(type) {
		case string:
			hostRes = append(hostRes, rt)
		case []string:
			hostRes = append(hostRes, rt...)
		default:
			log.Println("Func resultCheckSystemNetworkCardMAC:::Result type check Unknown")
		}
	}

	hostResRemoveRepeat := StringSliceRemoveRepeat(hostRes)

	hostResRepeat := DiffStringSlices(hostResRemoveRepeat, hostRes)
	if len(hostResRepeat) != 0 {
		log.Printf("Func resultCheckSystemNetworkCardMAC:::Host system network card exist repeat value %v, "+
			"Require network card non-repeat", hostResRepeat)
		os.Exit(1)
	}
}

func resultCheckSystemProductID(sl *SSHResultLog) {
	var (
		hostRes []string
	)
	hostRequired := checkHostResultValueFunc(sl, checkHostResultValueWorker)
	for _, resp := range hostRequired {
		switch rt := resp.Result.(type) {
		case string:
			hostRes = append(hostRes, rt)
		case []string:
			hostRes = append(hostRes, rt...)
		default:
			log.Println("Func resultCheckSystemProductID:::Result type check Unknown")
		}
	}

	hostResRemoveRepeat := StringSliceRemoveRepeat(hostRes)

	hostResRepeat := DiffStringSlices(hostResRemoveRepeat, hostRes)
	if len(hostResRepeat) != 0 {
		log.Printf("Func resultCheckSystemProductID:::Host system network card exist repeat value %v, "+
			"Require network card non-repeat", hostResRepeat)
		os.Exit(1)
	}
}

func resultCheckK8sPartNode(sl *SSHResultLog) {
	var hostNoRequire []HostRequired
	hostRequired := checkHostResultValueFunc(sl, checkHostResultValueWorker)
	for _, resp := range hostRequired {
		if resp.Result.(string) != "" {
			hostNoRequire = append(hostNoRequire, resp)
		}
	}
	if len(hostNoRequire) != 0 {
		for _, v := range hostNoRequire {
			log.Printf("Func resultCheckK8sPartNode:::Host %s system port %v occupied, "+
				"Require default port non-occupied", v.Host, v.Result)
		}
		os.Exit(1)
	}
}

func resultCheckSystemSwap(sl *SSHResultLog, requiredIntValue int) {
	var hostNoRequire []HostRequired
	requiredValue := strconv.Itoa(requiredIntValue)
	hostRequired := checkHostResultValueFunc(sl, checkHostResultValueWorker)
	for _, resp := range hostRequired {
		if resp.Result.(string) < requiredValue {
			hostNoRequire = append(hostNoRequire, resp)
		}
	}
	checkHostResultOutput(hostNoRequire, "Func resultCheckSystemSwap:::Host %s system swap space %v size, "+
		"show space Not closed,Require host system disable swap, swap space is equal to %d", requiredValue)
}

func resultCheckDockerInstall(sl *SSHResultLog, requireStringValue string) {
	var hostNoRequire []HostRequired
	hostRequired := checkHostResultValueFunc(sl, checkHostResultValueWorker)
	for _, resp := range hostRequired {
		if resp.Result.(string) == requireStringValue {
			hostNoRequire = append(hostNoRequire, resp)
		}
	}
	checkHostResultOutput(hostNoRequire,
		"Func resultCheckDockerInstall:::Host %s system docker Not install, Output %v-%v, "+
			"Require host system install docker", requireStringValue)
}

func resultCheckDockerVersion(sl *SSHResultLog, requireStringValue string) {
	var hostNoRequire []HostRequired
	hostRequired := checkHostResultValueFunc(sl, checkHostResultValueWorker)
	for _, resp := range hostRequired {
		if resp.Result.(string) < requireStringValue {
			hostNoRequire = append(hostNoRequire, resp)
		}
	}
	checkHostResultOutput(hostNoRequire,
		"Func resultCheckDockerVersion:::Host %s system docker version  %v non-compliant,"+
			"Require docker version is higher than or equal to %s", requireStringValue)
}

func resultCheckIpvsadmInstall(sl *SSHResultLog, requireStringValue string) {
	var hostNoRequire []HostRequired
	hostRequired := checkHostResultValueFunc(sl, checkHostResultValueWorker)
	for _, resp := range hostRequired {
		if resp.Result.(string) == "no" {
			hostNoRequire = append(hostNoRequire, resp)
		}
	}
	checkHostResultOutput(hostNoRequire,
		"Func resultCheckIpvsadmInstall:::Host %s system ipvsadm Not install, Output %v-%v, "+
			"Require host system install ipvsadm", requireStringValue)
}

/*
	Part: special function
*/
func checkSystemVersionWorker(res SSHResult, ch chan HostRequired) {
	var hostRequire HostRequired
	reg := regexp.MustCompile(`[0-9]+`)
	versionSlice := reg.FindAllString(res.Result, -1)
	if len(versionSlice) > 1 {
		version := fmt.Sprintf("%s.%s", versionSlice[0], versionSlice[1])
		v, err := strconv.ParseFloat(version, 64)
		if err != nil {
			log.Printf("Func resultCheckHostOSVersion:::Host %s OS version string convert float %v failed,"+
				"Error:%v\n", res.Host, version, err.Error())
		}
		hostRequire.Host = res.Host
		hostRequire.Result = v
	} else {
		version, err := strconv.ParseFloat(versionSlice[0], 64)
		if err != nil {
			log.Printf("Func resultCheckHostOSVersion:::Host %s OS version string convert float %v failed,"+
				"Error:%v\n", res.Host, version, err.Error())
		}
		hostRequire.Host = res.Host
		hostRequire.Result = version
	}
	ch <- hostRequire
}

func checkSystemNetworkWorker(res SSHResult, ch chan HostRequired) {
	var hostRequire HostRequired
	reg := regexp.MustCompile("([0-9.]+)[ ]*%")
	resp := reg.FindAllString(res.Result, -1)
	hostRequire.Host = res.Host
	hostRequire.Result = resp
	ch <- hostRequire
}

/*
	Part: common function
*/
func checkHostResultValueWorker(res SSHResult, ch chan HostRequired) {
	var hostRequire HostRequired
	hostRequire.Host = res.Host
	hostRequire.Result = res.Result
	ch <- hostRequire
}

func checkHostResultValueFunc(sl *SSHResultLog, fg func(r SSHResult, ch chan HostRequired)) []HostRequired {
	var (
		hostRequire []HostRequired
	)

	pool := NewPool(5, len(sl.SuccessHosts))
	chCheck := make([]chan HostRequired, len(sl.SuccessHosts))
	for i, res := range sl.SuccessHosts {
		chCheck[i] = make(chan HostRequired, 1)
		go func(r SSHResult, ch chan HostRequired) {
			pool.Acquire()
			fg(r, ch)
			pool.Release()
		}(res, chCheck[i])
		resp := <-chCheck[i]
		hostRequire = append(hostRequire, resp)
	}
	pool.Wg.Wait()
	return hostRequire
}

func checkHostResultOutput(hostNoRequire []HostRequired, errMsg string, requireValue string) {
	if len(hostNoRequire) != 0 {
		for _, v := range hostNoRequire {
			log.Printf(errMsg, v.Host, v.Result, requireValue)
		}
		os.Exit(1)
	}
}
