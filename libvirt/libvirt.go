package libvirt

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	"github.com/hashicorp/go-hclog"
	"libvirt.org/go/libvirt"
)

var (
	sshURIRegex = `^ssh:\/\/` + // Scheme
		`(([^@\/\s:]+)(:[^@\/\s:]+)?@)?` + // User (optional) and Password (optional)
		`([a-zA-Z0-9.\-]+)` + // Host
		`(:\d{1,5})?$` // Port (optional)

	ErrEmptyURI   = errors.New("connection URI can't be empty")
	ErrInvalidURI = errors.New("invalid connection URI")
)

type driver struct {
	uri    string
	conn   *libvirt.Connect
	logger hclog.Logger
}

func (d *driver) monitorCtx(ctx context.Context) {
	select {
	case <-ctx.Done():
		d.conn.Close()
		return
	}
}

func validURI(uri string) error {
	if uri == "" {
		return ErrEmptyURI
	}

	/* re := regexp.MustCompile(sshURIRegex)
	if !re.MatchString(uri) {
		return ErrInvalidURI
	} */

	return nil
}

func New(ctx context.Context, URI string, logger hclog.Logger) (*driver, error) {
	if err := validURI(URI); err != nil {
		return nil, err
	}

	conn, err := libvirt.NewConnect(URI)
	if err != nil {
		return nil, err
	}

	d := &driver{
		conn:   conn,
		logger: logger,
		uri:    URI,
	}

	go d.monitorCtx(ctx)

	return d, nil
}

func (d *driver) Close() (int, error) {
	return d.conn.Close()
}

func metadataAsString(m map[string]string) string {
	meta := []string{}
	for key, value := range m {
		meta = append(meta, fmt.Sprintf("%s=\"%s\"", key, value))
	}

	return strings.Join(meta, ",")
}

func (d *driver) parceVirtInstallArgs(dc *DomainConfig) []string {
	args := []string{
		fmt.Sprintf("--connect=%s", d.uri),
		fmt.Sprintf("--name=%s", dc.Name),
		fmt.Sprintf("--memory=%d", dc.Memory),
		fmt.Sprintf("--vcpus=%d", dc.CPUs),
		"--import",
		"--disk", "device=cdrom,path=/home/ubuntu/go/src/github.com/juanadelacuesta/heraclitus/vms/focal-server-cloudimg-amd64.img,format=qcow2",
		"--os-variant=ubuntu22.04",
		"--network", "bridge=virbr0,model=virtio",
		"--graphics", "vnc,listen=0.0.0.0",
		"--noautoconsole",
		"--cloud-init", "user-data=/home/ubuntu/go/src/github.com/juanadelacuesta/heraclitus/vms/user-data.yaml,meta-data=/home/ubuntu/go/src/github.com/juanadelacuesta/heraclitus/vms/meta-data.yaml,network-config=/home/ubuntu/go/src/github.com/juanadelacuesta/heraclitus/vms/network-config.yaml",
		"--print-xml=1",
	}

	//args = append(args, fmt.Sprintf("--id ", dc.ID))
	//args = append(args, fmt.Sprintf("--metadata %s", metadataAsString(dc.Metadata)))
	return args
}

func (d *driver) getXMLfromConfig(dc *DomainConfig) (string, error) {

	args := d.parceVirtInstallArgs(dc)
	cmd := exec.Command("virt-install", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (d *driver) CreateDomain(config *DomainConfig) {
	domainXML, err := d.getXMLfromConfig(config)
	if err != nil {
		fmt.Println("dolor", err)
		return
	}

	d.logger.Debug("define libvirt domain using xml: ", domainXML)
	dom, err := d.conn.DomainDefineXML(domainXML)
	if err != nil {
		fmt.Println("oh the error the second time", err)
	}
	fmt.Println(dom)
}

func (d *driver) createDomainXML() {
	config := &DomainConfig{
		Name: "blah",
	}

	tmpl := template.Must(template.New("domain").Parse(basicDomainXML))

	var domainXML bytes.Buffer
	if err := tmpl.Execute(&domainXML, config); err != nil {
		fmt.Println("oh the error", err)
	}

	d.logger.Info("define libvirt domain using xml: ", domainXML.String())
	dom, err := d.conn.DomainDefineXML(domainXML.String())
	fmt.Println("oh the error the second time", dom, err)
}

func (d *driver) GetVms() {
	doms, err := d.conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		fmt.Println("errores", err)
	}

	fmt.Printf("%d running domains:\n", len(doms))
	for _, dom := range doms {
		name, err := dom.GetName()
		if err == nil {
			fmt.Printf("  %s\n", name)
		}
		nam, err := dom.DomainGetConnect()
		if err == nil {
			fmt.Printf("  %+v\n", nam)
		}
		dom.Free()
	}
}
