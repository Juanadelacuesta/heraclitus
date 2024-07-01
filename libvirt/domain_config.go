package libvirt

import "fmt"

type DomainConfig struct {
	Name     string
	Image    string
	Metadata map[string]string
	Memory   int
	Cores    int
	CPUs     int
}

var configs map[string]option

func init() {
	configs = map[string]option{
		"name": {
			flag:      "name",
			mandatory: true,
		},
		"osinfo": {
			flag:       "osinfo",
			defaultVal: "detect=on,require=off",
		},
		"memory": {
			flag:       "memory",
			defaultVal: "1",
		},
	}
}

type option struct {
	flag       string
	mandatory  bool
	defaultVal string
}

func (o *option) value(v string) string {
	if v == "" && !o.mandatory {
		return ""
	}

	if v == "" {
		return fmt.Sprintf("--%s %s", o.flag, o.defaultVal)
	}

	return fmt.Sprintf("%s %s", o.flag, v)
}
