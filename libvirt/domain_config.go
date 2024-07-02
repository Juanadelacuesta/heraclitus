package libvirt

import "fmt"

/*type TaskConfig struct {
	ID               string
	JobName          string
	JobID            string
	TaskGroupName    string
	ParentJobID      string
	Name             string // task.Name
	Namespace        string
	NodeName         string
	NodeID           string
	Env              map[string]string
	DeviceEnv        map[string]string
	Resources        *Resources
	Devices          []*DeviceConfig
	Mounts           []*MountConfig
	User             string
	AllocDir         string
	rawDriverConfig  []byte
	StdoutPath       string
	StderrPath       string
	AllocID          string
	NetworkIsolation *NetworkIsolationSpec
	DNS              *DNSConfig
}*/

type resources struct {
}

type DomainConfig struct {
	Name         string
	Image        string
	Metadata     map[string]string
	Memory       int
	Cores        int
	CPUs         int
	OsVariant    string
	CloudImgPath string
	DiskFmt      string
}

func (d *driver) parceVirtInstallArgs(dc *DomainConfig) []string {
	args := []string{
		"--import",
		"--noautoconsole",
		"--print-xml=2",
		fmt.Sprintf("--connect=%s", d.uri),
		fmt.Sprintf("--name=%s", dc.Name),
		fmt.Sprintf("--ram=%d", dc.Memory),
		fmt.Sprintf("--vcpus=%d,cores=%d", dc.CPUs, dc.Cores),
		fmt.Sprintf("--os-variant=%s", dc.OsVariant),
		"--disk",
		fmt.Sprintf("path=%s,format=%s", dc.CloudImgPath, dc.DiskFmt),
		"--network", "bridge=virbr0,model=virtio",
		"--graphics", "vnc,listen=0.0.0.0",
		"--cloud-init", "user-data=/home/ubuntu/go/src/github.com/juanadelacuesta/heraclitus/vms/user-data.yaml,meta-data=/home/ubuntu/go/src/github.com/juanadelacuesta/heraclitus/vms/meta-data.yaml,network-config=/home/ubuntu/go/src/github.com/juanadelacuesta/heraclitus/vms/network-config.yaml",
	}

	//args = append(args, fmt.Sprintf("--id ", dc.ID))
	//args = append(args, fmt.Sprintf("--metadata %s", metadataAsString(dc.Metadata)))
	return args
}
