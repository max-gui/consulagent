package consulhelp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/max-gui/consulagent/pkg/consulsets"
	"github.com/max-gui/logagent/pkg/logagent"
	"gopkg.in/yaml.v2"
)

// var Consulurl = "http://dev-consul:8500"
// var AclToken string

// /**
//  * get app config template
//  */
// func getAppConfigTmplFromConsul(appName string) (string, error) {
// 	url := Consulurl + "/v1/kv/app/config/" + appName + "?token=" + AclToken
// 	log.Printf("get app config template url: %s", url)

// 	return httpGet(url)
// }
// init()

// var kvmap = make(map[string][]byte)

type mutexKV struct {
	sync.RWMutex
	kvs map[string]interface{}
}

var kvmap = mutexKV{kvs: make(map[string]interface{})}

func (v *mutexKV) help(tricky func(map[string]interface{}) (bool, interface{})) (bool, interface{}) {
	v.Lock()
	ok, res := tricky(v.kvs)
	v.Unlock()
	return ok, res
}

// func (v *mutexKV) readk(key string) ([]byte, bool) {
// 	v.Lock()
// 	if val, ok := v.kvs[key]; ok {
// 		return val, ok
// 	} else {
// 		return nil, ok
// 	}
// }

func ClsConfig() {
	kvmap.Lock()
	kvmap.kvs = make(map[string]interface{})
	kvmap.Unlock()
	// kvmap = make(map[string][]byte)
}

func StartWatch(prefix string, fulfil bool, c context.Context) {
	logger := logagent.Inst(c)
	watchConfig := make(map[string]interface{})
	watchConfig["type"] = "keyprefix"
	watchConfig["prefix"] = prefix
	// watchConfig["handler_type"] = "script"
	watchPlan, err := watch.Parse(watchConfig)
	watchPlan.Token = *consulsets.Acltoken
	if err != nil {
		logger.Panic(err)
	}
	// watchPlan.Type
	// 	watchPlan.Watcher = func(p *watch.Plan) (watch.BlockingParamVal, interface{}, error) {
	// 		p.HandlerType
	// 	}
	// var kvmap = make(map[string][]byte)
	watchPlan.Handler = func(lastIndex uint64, result interface{}) {

		keys := result.(api.KVPairs)
		// if keys == nil {
		// 	// vmap = make(map[string][]byte)
		// } else {
		// log.Print(watchPlan.Type)
		// log.Print(watchPlan.HandlerType)
		for _, v := range keys {
			if fulfil || v.ModifyIndex == lastIndex {
				logger.Print(string(v.Value))
				kvmap.help(func(kvs map[string]interface{}) (bool, interface{}) {
					kvs[v.Key] = v.Value
					return true, nil
				})
				// kvmap[v.Key] = v.Value
			}
		}
		// }
		fulfil = false

	}

	conf := api.DefaultConfig()
	conf.Address = *consulsets.Consul_host

	conf.Token = *consulsets.Acltoken

	err = watchPlan.Run(*consulsets.Consul_host)
	if err != nil {
		logger.Fatalf("start watch error, error message: %s", err.Error())
	}
}

/**
 * get resource information
 */
func Getconfaml(prefix, entityType, entityId, env string, c context.Context) map[string]interface{} {
	logger := logagent.Inst(c)
	maptmp := make(map[string]interface{})
	resbytes := Getconfibytes(prefix, entityType, entityId, env, c)
	err := yaml.Unmarshal(resbytes, &maptmp)
	if err != nil {
		logger.Panic(err)
	}

	return maptmp
}

func Getconfibytes(prefix, entityType, entityId, env string, c context.Context) []byte {
	maptmp := make(map[string]string)
	resbytes := getConfig(prefix, entityType, entityId, env, c)
	err := yaml.Unmarshal(resbytes, &maptmp)
	if err == nil {
		if idval, ok := maptmp["real-id"]; ok {
			if envval, ok := maptmp["real-env"]; ok {
				resbytes = getConfig(prefix, entityType, idval, envval, c)
			} else {
				resbytes = getConfig(prefix, entityType, idval, env, c)
			}
		}
	}
	return resbytes
}

// "ops/resource/" +
func DelConfig(prefix, entityType, entityId, env string, c context.Context) {
	var key = prefix + entityType + "/" + entityId + "/" + env

	DelConfigFull(key, c)
}

func DelConfigFull(key string, c context.Context) {
	// Get a new client
	logger := logagent.Inst(c)
	conf := api.DefaultConfig()
	conf.Address = *consulsets.Consul_host
	// conf.Address
	conf.Token = *consulsets.Acltoken
	client, err := api.NewClient(conf)
	if err != nil {
		panic(err)
	}

	writeOptions := &api.WriteOptions{Token: *consulsets.Acltoken}
	meta, err := client.KV().Delete(key, writeOptions)
	if err != nil {
		panic(err)
	}

	kvmap.help(func(kvs map[string]interface{}) (bool, interface{}) {
		delete(kvs, key)
		return true, nil
	})
	// delete(kvmap, key)

	logger.Print(meta)
}

func PutConfigFull(key string, value []byte, c context.Context) {
	logger := logagent.Inst(c)
	logger.Info(key)
	kv := &api.KVPair{
		Key:   key,
		Value: value,
	}
	// Get a new client
	conf := api.DefaultConfig()
	conf.Address = *consulsets.Consul_host
	// conf.Address
	conf.Token = *consulsets.Acltoken
	client, err := api.NewClient(conf)
	if err != nil {
		panic(err)
	}

	writeOptions := &api.WriteOptions{Token: *consulsets.Acltoken}
	meta, err := client.KV().Put(kv, writeOptions)
	if err != nil {
		panic(err)
	}
	logger.Info(meta)
}

func PutConfig(prefix, entityType, entityId, env string, value []byte, c context.Context) {
	var key = prefix + entityType + "/" + entityId + "/" + env
	PutConfigFull(key, value, c)
}

func GetConfigs(prefix, entityType string, c context.Context) api.KVPairs {
	logger := logagent.Inst(c)
	var key = prefix + entityType

	// Get a new client
	conf := api.DefaultConfig()
	conf.Address = *consulsets.Consul_host
	// conf.Address
	conf.Token = *consulsets.Acltoken
	client, err := api.NewClient(conf)
	if err != nil {
		panic(err)
	}

	// Get a handle to the KV API
	kv := client.KV()
	opts := &api.QueryOptions{Token: *consulsets.Acltoken}
	pairs, _, err := kv.List(key, opts)
	if err != nil {
		logger.Panic(err)
	}

	return pairs
}

func getConfig(prefix, entityType string, entityId string, env string, c context.Context) []byte {
	var key = prefix + entityType + "/" + entityId + "/" + env

	return GetConfigFull(key, c)
}

func GetServices(c context.Context) map[string][]string {
	// var key = prefix + entityType + "/" + entityId + "/" + env
	// log.Print(servicename)

	logger := logagent.Inst(c)
	conf := api.DefaultConfig()
	conf.Address = *consulsets.Consul_host
	// conf.Address
	conf.Token = *consulsets.Acltoken
	client, err := api.NewClient(conf)
	if err != nil {
		panic(err)
	}
	// Get a handle to the KV API
	catalog := client.Catalog()
	services, _, err := catalog.Services(&api.QueryOptions{})
	if err != nil {
		logger.Panic(err)
	}
	// log.Print(services)

	// res := map[string]map[string]string{}
	// for k,v :=range services{
	// 	res[k] =
	// }

	return services
	// // PUT a new KV pair
	// p := &api.KVPair{Key: "REDIS_MAXCLIENTS", Value: []byte("1000")}
	// _, err = kv.Put(p, nil)
	// if err != nil {
	// 	panic(err)
	// }

	// Lookup the pair

}

func GetService(servicename string, c context.Context) []*api.CatalogService {
	// var key = prefix + entityType + "/" + entityId + "/" + env
	logger := logagent.Inst(c)
	logger.Info(servicename)

	conf := api.DefaultConfig()
	conf.Address = *consulsets.Consul_host
	// conf.Address
	conf.Token = *consulsets.Acltoken
	client, err := api.NewClient(conf)
	if err != nil {
		panic(err)
	}
	// Get a handle to the KV API
	catalog := client.Catalog()
	service, _, err := catalog.Service(servicename, "", &api.QueryOptions{})
	if err != nil {
		logger.Panic(err)
	}
	logger.Info(service)

	return service
	// // PUT a new KV pair
	// p := &api.KVPair{Key: "REDIS_MAXCLIENTS", Value: []byte("1000")}
	// _, err = kv.Put(p, nil)
	// if err != nil {
	// 	panic(err)
	// }

	// Lookup the pair

}

type serviceInfo struct {
	serviceentry []*api.ServiceEntry
	lastCheck    time.Time
}

func GetHealthService(servicename string, c context.Context) []*api.ServiceEntry {
	// var key = prefix + entityType + "/" + entityId + "/" + env
	logger := logagent.Inst(c)
	logger.Info(servicename)
	// time.Since(time.Now()).Minutes()
	// f := *consulsets.Cacheminutes * time.Now().Minute()
	if ok, value := kvmap.help(func(kvs map[string]interface{}) (bool, interface{}) {
		if val, ok := kvs[servicename]; ok {
			return ok, val
		} else {
			return ok, nil
		}
	}); ok {

		if time.Duration(*consulsets.Cacheminutes)*time.Minute > time.Since(value.(serviceInfo).lastCheck) {
			return value.(serviceInfo).serviceentry
		}
	}

	conf := api.DefaultConfig()
	conf.Address = *consulsets.Consul_host
	// conf.Address
	conf.Token = *consulsets.Acltoken
	client, err := api.NewClient(conf)
	if err != nil {
		panic(err)
	}
	// Get a handle to the KV API
	health := client.Health()
	service, _, err := health.Service(servicename, "", true, &api.QueryOptions{})
	if err != nil {
		logger.Panic(err)
	}
	logger.Info(service)

	// var bytetmps bytes.Buffer
	// enc := gob.NewEncoder(&bytetmps)
	// err = enc.Encode(service)
	// if err != nil {
	// 	logger.Panic(err)
	// }
	kvmap.help(func(kvs map[string]interface{}) (bool, interface{}) {

		kvs[servicename] = serviceInfo{serviceentry: service, lastCheck: time.Now()}
		return true, nil
	})

	return service
	// // PUT a new KV pair
	// p := &api.KVPair{Key: "REDIS_MAXCLIENTS", Value: []byte("1000")}
	// _, err = kv.Put(p, nil)
	// if err != nil {
	// 	panic(err)
	// }

	// Lookup the pair

}

func GetConfigFull(key string, c context.Context) []byte {
	// var key = prefix + entityType + "/" + entityId + "/" + env
	logger := logagent.Inst(c)
	logger.Info(key)

	// if val, ok := kvmap[key]; ok {
	// 	return val
	// }

	if ok, value := kvmap.help(func(kvs map[string]interface{}) (bool, interface{}) {
		if val, ok := kvs[key]; ok {
			return ok, val
		} else {
			return ok, nil
		}
	}); ok {
		return value.([]byte)
	}

	// if val, ok := readkv(key); ok {
	// 	return val
	// }

	// rediscli := redisops.Pool().Get()

	// defer rediscli.Close()

	// bytes, err := redis.Bytes(rediscli.Do("GET", key))
	// if err == nil && len(bytes) > 0 {
	// 	rediscli.Do("SETEX", key, 6000, bytes)
	// 	return bytes
	// }

	// Get a new client
	conf := api.DefaultConfig()
	conf.Address = *consulsets.Consul_host
	// conf.Address
	conf.Token = *consulsets.Acltoken
	client, err := api.NewClient(conf)
	if err != nil {
		logger.Panic(err)
	}

	// Get a handle to the KV API
	kv := client.KV()

	// // PUT a new KV pair
	// p := &api.KVPair{Key: "REDIS_MAXCLIENTS", Value: []byte("1000")}
	// _, err = kv.Put(p, nil)
	// if err != nil {
	// 	panic(err)
	// }

	// Lookup the pair
	pair, _, err := kv.Get(key, nil)
	if err != nil {
		logger.Panic(err)
	}

	if pair == nil {
		logger.Panicf("dont have the key:%s", key)
		return nil
	}

	logger.Infof("KV key: %v\n", pair.Key)
	logger.Infof("KV value: %s\n", string(pair.Value))
	// re, err := rediscli.Do("SETEX", pair.Key, 6000, pair.Value)

	// kvmap[key] = pair.Value

	kvmap.help(func(kvs map[string]interface{}) (bool, interface{}) {
		kvs[key] = pair.Value
		return true, nil
	})

	return pair.Value
}

// /**
//  * get resource information
//  */
// func GetInfrastructureInfo(entityType string, entityId string, env string) (string, error) {
// 	var url = *consulsets.Consul_host + "/v1/kv/ops/resource/" + entityType + "/" + entityId + "/" + env
// 	if len(*consulsets.Acltoken) > 0 {
// 		url += "?token=" + *consulsets.Acltoken
// 	}

// 	log.Printf("get infrastructure info url: %s", url)

// 	return httpGet(url)
// }

/**
 * get resource information
 */
func Sendconfig2consul(entityType string, entityId string, env string, content string) (string, error) {
	var url = *consulsets.Consul_host + "/v1/kv/ops/resource/" + entityType + "/" + entityId + "/" + env
	if len(*consulsets.Acltoken) > 0 {
		url += "?token=" + *consulsets.Acltoken
	}
	log.Printf("get infrastructure info url: %s", url)

	return httpPut(url, content)
}

/**
 * trigger http put request
 */
func httpPut(url string, body string) (string, error) {
	payload := strings.NewReader(body)

	req, reqerr := http.NewRequest("PUT", url, payload)
	if reqerr != nil {
		return "", reqerr
	} else {
		res, reserr := http.DefaultClient.Do(req)

		if reserr != nil {
			return "", reserr
		} else {
			defer res.Body.Close()
			resbody, resbodyerr := ioutil.ReadAll(res.Body)

			if resbodyerr != nil {
				return "", resbodyerr
			} else {
				return string(resbody), resbodyerr
			}
		}
	}

}

/**
???json????????????????????????map??????
*/
func extractValueFromJsonMsg(jsonString string) string {
	var list []map[string]interface{}

	err := json.Unmarshal([]byte(jsonString), &list)
	if err != nil {
		log.Fatalf("convert json to map error: %v", err)
	}

	return fmt.Sprintf("%v", list[0]["Value"])
	// return convertops.StrValOfInterface(list[0]["Value"])
}

func base64Decode(value string) string {
	result, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		fmt.Printf("base64 decode failure, error=[%v]\n", err)
	}
	return string(result)
}
