package f5

import (
	"strings"
)

type Interface struct {
	Name         string      `json:"name"`
	FullPath     string      `json:"fullPath"`
	Generation   int         `json:"generation"`
	MacAddress   string      `json:"macAddress"`
	MTU          int         `json:"mtu"`
	MediaMax     string      `json:"mediaMax"`
	MediaActive  string      `json:"mediaActive"`
}

type Interfaces struct {
	Items []Interface `json:"items"`
}

type InterfaceStatsDescription struct {
	Description string `json:"description"`
}

type InterfaceStatsInnerEntries struct {
	PktsOut         LBStatsValue           `json:"counters.pktsOut"`
	PktsIn          LBStatsValue           `json:"counters.pktsIn"`
	ErrorsAll       LBStatsValue           `json:"counters.errorsAll"`
	DropsAll        LBStatsValue           `json:"counters.dropsAll"`
	Status          InterfaceStatsDescription `json:"status"`
	MediaActive     InterfaceStatsDescription `json:"mediaActive"`
	TmNape          InterfaceStatsDescription `json:"tmName"`
	BitsIn          LBStatsValue           `json:"counters.bitsIn"`
	BitsOut         LBStatsValue           `json:"counters.bitsOut"`
}

type InterfaceStatsNestedStats struct {
	Kind     string                  `json:"kind"`
	SelfLink string                  `json:"selfLink"`
	Entries  InterfaceStatsInnerEntries `json:"entries"`
}

type InterfaceURLKey struct {
	NestedStats InterfaceStatsNestedStats `json:"nestedStats"`
}
type InterfaceStatsOuterEntries map[string]InterfaceURLKey

type InterfaceStats struct {
	Kind       string                  `json:"kind"`
	SelfLink   string                  `json:"selfLink"`
	Entries    InterfaceStatsOuterEntries `json:"entries"`
}

func (f *Device) ShowInterfaces() (error, *Interfaces) {

	u := f.Proto + "://" + f.Hostname + "/mgmt/tm/net/interface"
	res := Interfaces{}

	err, _ := f.sendRequest(u, GET, nil, &res)
	if err != nil {
		return err, nil
	} else {
		return nil, &res
	}

}

func (f *Device) ShowInterface(rname string) (error, *Interface) {

	iface := strings.Replace(rname, "/", "~", -1)
	u := f.Proto + "://" + f.Hostname + "/mgmt/tm/net/interface/" + iface
	res := Interface{}

	err, _ := f.sendRequest(u, GET, nil, &res)
	if err != nil {
		return err, nil
	} else {
		return nil, &res
	}

}

func (f *Device) ShowInterfaceStats(rname string) (error, *LBObjectStats) {

	iface := strings.Replace(rname, "/", "~", -1)
	u := f.Proto + "://" + f.Hostname + "/mgmt/tm/net/interface/" + iface + "/stats"
	res := LBObjectStats{}

	err, _ := f.sendRequest(u, GET, nil, &res)
	if err != nil {
		return err, nil
	} else {
		return nil, &res
	}

}

func (f *Device) ShowAllInterfaceStats() (error, *InterfaceStats) {

	u := f.Proto + "://" + f.Hostname + "/mgmt/tm/net/interface/stats"
	res := InterfaceStats{}

	err, _ := f.sendRequest(u, GET, nil, &res)
	if err != nil {
		return err, nil
	} else {
		return nil, &res
	}

}
