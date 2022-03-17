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
	Acltoken = flag.String("acltoken", "", "consul acltoken")

	Consul_host = flag.String("consulhost", "http://consul-stg.kube.com", "consul url") //http://consul-prod.kube.com
	Cacheminutes = flag.Int("cacheminutes", 5, "service cache for minutes")
}

// var Reppath = func() string {
// 	return *Apppath + PthSep + *Repopathname + PthSep
// }
