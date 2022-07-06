package consulsets

import (
	"flag"
)

var (
	Acltoken, Consul_host *string
	Cacheminutes          *int
)

func StartupInit(acltoken string) {

	// confmap := map[string]interface{}{}
	// yaml.Unmarshal(bytes, confmap)
	*Acltoken = acltoken // confmap["af-arch"].(map[interface{}]interface{})["resource"].(map[interface{}]interface{})["private"].(map[interface{}]interface{})["acl-token"].(string)
}

func init() {
	// Appname = "charon"
	Acltoken = flag.String("acltoken", "13a3dbe0-199c-af04-ae9d-b4c43eb735eb", "consul acltoken") //245d0a09-7139-config-prod-ff170a0562b1

	Consul_host = flag.String("consulhost", "http://consul-stg.paic.com.cn", "consul url") //http://consul-prod.kube.com http://consul-stg.kube.com http://consul-szf-prod.kube.com
	Cacheminutes = flag.Int("cacheminutes", 5, "service cache for minutes")
}

// var Reppath = func() string {
// 	return *Apppath + PthSep + *Repopathname + PthSep
// }
