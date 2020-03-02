package network

import (
	"fmt"
	"testing"
)

func TestNewNetwork(t *testing.T) {
	netYaml := NewNetwork("calico", MetaData{Interface: "en.*|eth.*", PodCIDR: "10.1.1.1/24"}).Manifests("")
	fmt.Println(netYaml)
}
