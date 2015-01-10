package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/jmcvetta/napping"
	//	"github.com/kr/pretty"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"net/url"
)

var (
	f5Host      string
	username    string
	passwd      string
	credentials map[string]string
	debug       bool = false
	partition   string
	poolname    string
	cfgFile     string = "f5.json"
)

/*
var f5Cmd = &cobra.Command{
	Use:   "f5er",
	Short: "tickle an F5 load balancer using REST",
	Long:  "A utility to create and manage F5 configuration objects",
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "show F5 objects",
	Long:  "show current state of F5 objects. Show requires an object, eg. f5er show pool",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			bail("show needs an argument - the object type to show perhaps??")
		}
	},
}

var showPoolCmd = &cobra.Command{
	Use:   "pool",
	Short: "show a pool",
	Long:  "show the current state of a pool",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("f5er show pool")
			showPools()
		} else {
			pname := args[0]
			showPool(pname)
		}
	},
}
*/

type httperr struct {
	Message string
	Errors  []struct {
		Resource string
		Field    string
		Code     string
	}
}

/*
"membersReference":{
				"link":"https://localhost/mgmt/tm/ltm/pool/~DMZ~audmzbilltweb-sit_443_pool/members?ver=11.6.0",
				"isSubcollection":true,
				"items":[
				  {
									"kind":"tm:ltm:pool:members:membersstate",
									"name":"audmzbilltweb01-sit:443",
									"partition":"DMZ",
									"fullPath":"/DMZ/audmzbilltweb01-sit:443",
									"generation":233,
									"selfLink":"https://localhost/mgmt/tm/ltm/pool/~DMZ~audmzbilltweb-sit_443_pool/members/~DMZ~audmzbilltweb01-sit:443?ver=11.6.0",
									"address":"10.60.61.215%6",
									"connectionLimit":0,
									"dynamicRatio":1,
									"ephemeral":"false",
									"fqdn":{
													"autopopulate":"disabled"
									},
									"inheritProfile":"enabled",
									"logging":"disabled",
									"monitor":"default",
									"priorityGroup":0,
									"rateLimit":"disabled",
									"ratio":1,
									"session":"monitor-enabled",
									"state":"up"
					}
			]
	}
*/

type LBMember struct {
	Name     string `json:"name"`
	Fullpath string `json:"fullPath"`
	Address  string `json:"address"`
	State    string `json:"state"`
}
type LBMemberRef struct {
	Link  string     `json:"link"`
	Items []LBMember `json":items"`
}

type LBPool struct {
	Name              string      `json:"name"`
	Fullpath          string      `json:"fullPath"`
	Generation        int         `json:"generation"`
	AllowNat          string      `json:"allowNat"`
	AllowSnat         string      `json:"allowSnat"`
	LoadBalancingMode string      `json:"loadBalancingMode"`
	Monitor           string      `json:"monitor"`
	MemberRef         LBMemberRef `json:"membersReference"`
}

/*
{"kind":"tm:ltm:pool:poolstate","name":"audmzbilltweb-sit_443_pool","partition":"DMZ","fullPath":"/DMZ/audmzbilltweb-sit_443_pool","generation":233,"selfLink":"https://localhost/mgmt/tm/ltm/pool/~DMZ~audmzbilltweb-sit_443_pool?ver=11.6.0","allowNat":"yes","allowSnat":"yes","ignorePersistedWeight":"disabled","ipTosToClient":"pass-through","ipTosToServer":"pass-through","linkQosToClient":"pass-through","linkQosToServer":"pass-through","loadBalancingMode":"round-robin","minActiveMembers":0,"minUpMembers":0,"minUpMembersAction":"failover","minUpMembersChecking":"disabled","monitor":"min 1 of { /Common/tcp }","queueDepthLimit":0,"queueOnConnectionLimit":"disabled","queueTimeLimit":0,"reselectTries":0,"serviceDownAction":"none","slowRampTime":10,"membersReference":{"link":"https://localhost/mgmt/tm/ltm/pool/~DMZ~audmzbilltweb-sit_443_pool/members?ver=11.6.0","isSubcollection":true}}
*/

type LBPools struct {
	Items []LBPool `json:"items"`
}

func InitialiseConfig() {
	viper.SetConfigFile(cfgFile)
	viper.AddConfigPath(".")
	if f5Host != "" {
		viper.Set("f5", f5Host)
	}
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Can't find your config file: %s\n", cfgFile)
	}
	if !viper.IsSet("credentials") {
		log.Fatal("no login credentials defined in config")
	}
	credentials = viper.GetStringMapString("credentials")
	var ok bool
	username, ok = credentials["username"]
	if !ok {
		log.Fatal("no username defined in config")
	}
	passwd, ok = credentials["passwd"]
	if !ok {
		log.Fatal("no passwd defined in config")
	}
	checkRequiredFlag("f5")

	f5Host = viper.GetString("f5")
	partition = viper.GetString("partition")
	debug = viper.GetBool("debug")
	poolname = viper.GetString("poolname")
}

func checkRequiredFlag(flg string) {
	if !viper.IsSet(flg) {
		log.SetFlags(0)
		log.Fatalf("\nerror: missing required option --%s\n\n", flg)
	}
}

func bail(msg string) {
	log.SetFlags(0)
	log.Fatalf("\n%s\n\n", msg)
}

func showPools() {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	//
	// Setup HTTP Basic auth for this session (ONLY use this with SSL).  Auth
	// can also be configured on a per-request basis when using Send().
	//
	s := napping.Session{
		Client:   client,
		Log:      debug,
		Userinfo: url.UserPassword(username, passwd),
	}

	url := "https://" + f5Host + "/mgmt/tm/ltm/pool"
	res := LBPools{}

	//url := "https://10.60.99.241/mgmt/tm/ltm/pool/~DMZ~audmzbilltweb-sit_443_pool"
	//url := "https://10.60.99.241/mgmt/tm/ltm/pool"
	//url := "https://10.60.99.242/mgmt/tm/ltm/pool/~DMZ~audmzbilltweb-sit_443_pool/members?ver=11.6.0"
	//
	// Send request to server
	//
	e := httperr{}
	resp, err := s.Get(url, nil, &res, &e)
	if err != nil {
		log.Fatal(err)
	}
	if resp.Status() == 401 {
		log.Fatal("unauthorised - check your username and passwd")
	}
	if resp.Status() >= 300 {
		log.Fatal(e.Message)
	}
	for _, v := range res.Items {
		fmt.Printf("pool:\t%s\n", v.Fullpath)
	}
	/*
		if debug {
			fmt.Printf("url:\t%s\n\n", url)
			fmt.Println("response Status:", resp.Status())
			fmt.Println("--------------------------------------------------------------------------------")
			fmt.Println("Header")
			fmt.Println(resp.HttpResponse().Header)
			fmt.Println("--------------------------------------------------------------------------------")
			fmt.Println("RawText")
			fmt.Println(resp.RawText())
			println("")
		}
	*/
}

func showPool(pname string) {

	/*
		{
						"kind":"tm:ltm:pool:poolstate",
						"name":"audmzbilltweb-sit_443_pool",
						"partition":"DMZ",
						"fullPath":"/DMZ/audmzbilltweb-sit_443_pool",
						"generation":156,
						"selfLink":"https://localhost/mgmt/tm/ltm/pool/~DMZ~audmzbilltweb-sit_443_pool?ver=11.6.0",
						"allowNat":"yes",
						"allowSnat":"yes",
						"ignorePersistedWeight":"disabled",
						"ipTosToClient":"pass-through",
						"ipTosToServer":"pass-through",
						"linkQosToClient":"pass-through",
						"linkQosToServer":"pass-through",
						"loadBalancingMode":"round-robin",
						"minActiveMembers":0,
						"minUpMembers":0,
						"minUpMembersAction":"failover",
						"minUpMembersChecking":"disabled",
						"monitor":"/Common/https ",
						"queueDepthLimit":0,
						"queueOnConnectionLimit":"disabled",
						"queueTimeLimit":0,
						"reselectTries":0,
						"serviceDownAction":"none",
						"slowRampTime":10,
						"membersReference":{
										"link":"https://localhost/mgmt/tm/ltm/pool/~DMZ~audmzbilltweb-sit_443_pool/members?ver=11.6.0",
										"isSubcollection":true
						}
		}
	*/

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	//
	// Setup HTTP Basic auth for this session (ONLY use this with SSL).  Auth
	// can also be configured on a per-request basis when using Send().
	//
	s := napping.Session{
		Client:   client,
		Log:      debug,
		Userinfo: url.UserPassword(username, passwd),
	}

	//	url := "https://" + f5Host + "/mgmt/tm/ltm/pool/" + poolname + "?\\$expand=*"
	//url := "https://" + f5Host + "/mgmt/tm/ltm/pool/~" + partition + "~" + poolname + "~expand=*"
	url := "https://" + f5Host + "/mgmt/tm/ltm/pool/~" + partition + "~" + poolname + "?expandSubcollections=true"
	res := LBPool{}

	//url := "https://10.60.99.241/mgmt/tm/ltm/pool/~DMZ~audmzbilltweb-sit_443_pool"
	//url := "https://10.60.99.241/mgmt/tm/ltm/pool"
	//url := "https://10.60.99.242/mgmt/tm/ltm/pool/~DMZ~audmzbilltweb-sit_443_pool/members?ver=11.6.0"
	//
	// Send request to server
	//
	e := httperr{}
	resp, err := s.Get(url, nil, &res, &e)
	if err != nil {
		log.Fatal(err)
	}
	if resp.Status() == 401 {
		log.Fatal("unauthorised - check your username and passwd")
	}
	if resp.Status() >= 300 {
		log.Fatal(e.Message)
	}
	fmt.Printf("pool name:\t%s\n", res.Name)
	fmt.Printf("fullpath:\t%s\n", res.Fullpath)
	fmt.Printf("lb mode:\t%s\n", res.LoadBalancingMode)
	fmt.Printf("monitor:\t%s\n", res.Monitor)

	for i, member := range res.MemberRef.Items {
		fmt.Printf("\tmember %d name:\t\t%s\n", i, member.Name)
		fmt.Printf("\tmember %d address:\t%s\n", i, member.Address)
		fmt.Printf("\tmember %d state:\t\t%s\n", i, member.State)
	}
	/*
		if debug {
			fmt.Printf("url:\t%s\n\n", url)
			fmt.Println("response Status:", resp.Status())
			fmt.Println("--------------------------------------------------------------------------------")
			fmt.Println("Header")
			fmt.Println(resp.HttpResponse().Header)
			fmt.Println("--------------------------------------------------------------------------------")
			fmt.Println("Response")
			jsonresp, err := prettifyJson(resp.RawText())
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(jsonresp)
			fmt.Println("--------------------------------------------------------------------------------")
		}
	*/
}

func prettifyJson(s string) (string, error) {
	var data map[string]interface{}
	b := []byte(s)
	if err := json.Unmarshal(b, &data); err != nil {
		return s, err
	}
	b, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return s, err
	}
	return string(b), nil
}

func init() {
	f5Cmd.Flags().StringVarP(&f5Host, "f5", "f", "", "IP or hostname of F5 to poke")
	f5Cmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug output")
	f5Cmd.PersistentFlags().StringVarP(&partition, "partition", "p", "", "F5 partition")
	viper.BindPFlag("f5", f5Cmd.Flags().Lookup("f5"))
	viper.BindPFlag("debug", f5Cmd.Flags().Lookup("debug"))
	viper.BindPFlag("partition", f5Cmd.Flags().Lookup("partition"))
	f5Cmd.AddCommand(showCmd)
	showCmd.AddCommand(showPoolCmd)
	//	log.SetFlags(log.Ltime | log.Lshortfile)
	log.SetFlags(0)
	InitialiseConfig()
}

func main() {
	viper.AutomaticEnv()
	f5Cmd.Execute()
}
