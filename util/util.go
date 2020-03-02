package util

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/daviddengcn/go-colortext"
)

const fileChunk = 8192 // we settle for 8KB

func Md5SumLocalFile(fileName string) string {
	file, err := os.Open(fileName)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("File %s Md5Sum Failed,Error: %v\n", fileName, err)
		}
	}()
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()

	// calculate the file size
	info, _ := file.Stat()
	fileSize := info.Size()
	blocks := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))

	hash := md5.New()
	for i := uint64(0); i < blocks; i++ {
		blocksize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
		buf := make([]byte, blocksize)
		file.Read(buf)
		io.WriteString(hash, string(buf)) // append into the hash
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// determine if the given path file / folder exists
func FileExists(path string) bool {
	// os.Stat get file info
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// StringSliceCountValues function,find out how many times there are duplicate elements in an array and their values,
// If the number of slice elements is higher than 1, the array have duplicate elements
func StringSliceCountValues(args []string) (Status bool, AllRepeatValue []string, MaxCount int,
	MaxValue []string) {
	// Exit without value
	if len(args) == 0 {
		return false, nil, 0, nil
	}

	// Find the number of occurrences corresponding to each value, for example: [value: times, value: times]
	newMap := make(map[string]int)
	for _, value := range args {
		if newMap[value] != 0 {
			newMap[value]++
		} else {
			newMap[value] = 1
		}
	}

	// Find the most occurrences
	var (
		allCount    []int    // all times
		maxCount    int      // most occurrences
		repeatValue []string // values that occur more than 1
	)
	for key, value := range newMap {
		// find the value of occurrences greater than 1
		if value > 1 {
			repeatValue = append(repeatValue, key)
		}
		allCount = append(allCount, value)
	}
	maxCount = allCount[0]
	for i := 0; i < len(allCount); i++ {
		if maxCount < allCount[i] {
			maxCount = allCount[i]
		}
	}

	// Find the most frequently occurring value in the array, for example: [8,9] These two values appear as many times as
	var maxValue []string
	for key, value := range newMap {
		if value == maxCount {
			maxValue = append(maxValue, key)
		}
	}

	return true, repeatValue, maxCount, maxValue
}

// StringTwoSliceRepeatElem function,find two slice repeat elem
func StringTwoSliceRepeatElem(a, b []string) []string {
	var u []string
	for _, v := range a {
		if stringsContains(b, v) {
			u = append(u, v)
		}
	}
	return u
}

func stringsContains(array []string, val string) bool {
	for i := 0; i < len(array); i++ {
		if array[i] == val {
			return true
		}
	}
	return false
}

// CheckAndSplitCommandArgs function,check one and more command params whether flagFormat correct,for example: hostname;date;ip addr
func CheckAndSplitCommandArgs(cmdArgs string) (cmdList []string) {
	if cmdArgs != "" {
		if !strings.Contains(cmdArgs, ",") {
			cmdList = append(cmdList, cmdArgs)
		}
		cmdList = strings.Split(cmdArgs, ",")
	}
	return cmdList
}

// StringSliceRemoveRepeat is used for single string slice remove repeat
func StringSliceRemoveRepeat(stringSliceList []string) []string {
	mapList := make(map[string]interface{})
	if len(stringSliceList) <= 0 {
		return nil
	}
	for _, v := range stringSliceList {
		mapList[v] = 1
	}
	var dataSlice []string
	for k := range mapList {
		if k == "" {
			continue
		}
		dataSlice = append(dataSlice, k)
	}
	return dataSlice
}

// DiffStringSlices function,Get the different items between two slices
// diffStringSlices function,Get difference items between two slices
func DiffStringSlices(slice1 []string, slice2 []string) []string {
	var diffStr []string
	m := map[string]int{}
	for _, s1Val := range slice1 {
		m[s1Val] = 1
	}
	for _, s2Val := range slice2 {
		m[s2Val] = m[s2Val] + 1
	}
	for mKey, mVal := range m {
		if mVal == 1 {
			diffStr = append(diffStr, mKey)
		}
	}
	return diffStr
}

// ColorPrintWithTerminalStyle function, console output color control, compatible with Windows & Linux
func ColorPrintWithTerminalStyle(logLevel string, textBefore interface{}, colorText string,
	textAfter ...interface{}) {
	color := ct.None
	switch logLevel {
	case "INFO":
		color = ct.Green
	case "WARN":
		color = ct.Yellow
	case "ERROR":
		color = ct.Red
	case "Task":
		color = ct.Cyan

	}

	fmt.Printf("%s", textBefore)
	ct.Foreground(color, true)
	fmt.Printf("%s", colorText)
	ct.ResetColor()
	for _, v := range textAfter {
		fmt.Printf("%s", v)
	}
}

// ParseIPSegment，for example IP 192.168.0.2-192.168.0.6
func ParseIPSegment(ips []string) []string {
	var hosts []string
	for _, nodes := range ips {
		if len(nodes) > 15 {
			log.Println("Multi-master/Multi-worker illegal.")
			os.Exit(1)
		} else if !strings.Contains(nodes, "-") {
			hosts = append(hosts, nodes)
			continue
		} else {
			startIP := strings.Split(nodes, "-")[0]
			endIP := strings.Split(nodes, "-")[1]
			ipPos := strings.LastIndex(nodes, ".")
			endIP = nodes[:ipPos+1] + endIP
			hosts = append(hosts, startIP)
			for Cmp(stringToIP(startIP), stringToIP(endIP)) < 0 {
				startIP = NextIP(stringToIP(startIP)).String()
				hosts = append(hosts, startIP)
			}
		}

	}
	return hosts
}

// Cmp compares two IPs, returning the usual ordering:
// a < b : -1
// a == b : 0
// a > b : 1
func Cmp(a, b net.IP) int {
	aa := ipToInt(a)
	bb := ipToInt(b)
	return aa.Cmp(bb)
}

func ipToInt(ip net.IP) *big.Int {
	if v := ip.To4(); v != nil {
		return big.NewInt(0).SetBytes(v)
	}
	return big.NewInt(0).SetBytes(ip.To16())
}

func intToIP(i *big.Int) net.IP {
	return net.IP(i.Bytes())
}

func stringToIP(i string) net.IP {
	return net.ParseIP(i).To4()
}

// NextIP returns IP incremented by 1
func NextIP(ip net.IP) net.IP {
	i := ipToInt(ip)
	return intToIP(i.Add(i, big.NewInt(1)))
}

// Kubernentes version convert to int
func K8sVersionConvertToInt(version string) int {
	// v1.15.6  => 1.15.6
	version = strings.Replace(version, "v", "", -1)
	versionArr := strings.Split(version, ".")
	if len(versionArr) >= 2 {
		versionStr := versionArr[0] + versionArr[1]
		if i, err := strconv.Atoi(versionStr); err == nil {
			return i
		}
	}
	return 0
}
