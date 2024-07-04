package libvirt

import (
	"fmt"
)

type Users struct {
	Default bool
	Users   []UserConfig
}

type UserConfig struct {
	Name     string
	Password string
	SSHKeys  []string
	Sudo     string
	Groups   []string
	Shell    string
}

type cloudinitConfig struct {
	metadataPath string
	userDataPath string
}

type DomainConfig struct {
	Name             string
	Image            string
	Metadata         map[string]string
	Memory           int
	Cores            int
	CPUs             int
	OsVariant        string
	CloudImgPath     string
	DiskFmt          string
	NetworkInterface string
	Type             string
	HostName         string
	UsersConfig      Users
	EnvVariables     map[string]string
}

func (d *driver) parceVirtInstallArgs(dc *DomainConfig, ci *cloudinitConfig) []string {
	args := []string{
		"--debug",
		fmt.Sprintf("--connect=%s", d.uri),
		fmt.Sprintf("--name=%s", dc.Name),
		fmt.Sprintf("--ram=%d", dc.Memory),
		fmt.Sprintf("--vcpus=%d,cores=%d", dc.CPUs, dc.Cores),
		fmt.Sprintf("--os-variant=%s", dc.OsVariant),
		"--import", "--disk", fmt.Sprintf("path=%s,format=%s", dc.CloudImgPath, dc.DiskFmt),
		"--network", fmt.Sprintf("bridge=%s,model=virtio", dc.NetworkInterface),
		"--graphics", "vnc,listen=0.0.0.0",
		"--cloud-init", fmt.Sprintf("user-data=%s,meta-data=%s,disable=on", ci.userDataPath, ci.metadataPath),
		"--noautoconsole",
	}
	return args
}
