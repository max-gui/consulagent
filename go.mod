module github.com/max-gui/consulagent

go 1.15

require (
	github.com/hashicorp/consul/api v1.9.1
	github.com/max-gui/logagent v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v2 v2.2.8

)

replace github.com/max-gui/logagent => ../logagent
