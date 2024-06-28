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
	ErrEmptyURI = errors.New("connection URI can't be empty")
)

type driver struct {
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

func New(ctx context.Context, URI string, logger hclog.Logger) (*driver, error) {
	if URI == "" {
		return nil, ErrEmptyURI
	}

	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		return nil, err
	}
	d := &driver{
		conn:   conn,
		logger: logger,
	}

	go d.monitorCtx(ctx)

	return d, nil
}

func (d *driver) Close() (int, error) {
	return d.conn.Close()
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

	d.createDomain()

	fmt.Println("   running again")
	doms, err = d.conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
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

type domainConfig struct {
	Name     string
	Image    string
	ID       string
	Metadata map[string]string
}

func metadataAsString(m map[string]string) string {
	meta := []string{}
	for key, value := range m {
		meta = append(meta, fmt.Sprintf("%s=\"%s\"", key, value))
	}

	return strings.Join(meta, ",")
}

func parceVirtInstallArgs(dc domainConfig) []string {

	args := []string{}
	args = append(args, "--check all")
	args = append(args, fmt.Sprintf("--name %s", dc.Name))
	args = append(args, fmt.Sprintf("--id "))
	return args
}

func (d *driver) formXMLfromConfig(domainConfig) (bytes.Buffer, error) {
	var outb, errb bytes.Buffer

	cmd := exec.Command("ls", "/usr/local/bin")
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil {
		return bytes.Buffer{}, err
	}

	return outb, nil
}

func (d *driver) createDomain() {
	config := &domainConfig{
		Name: "blah",
		ID:   "16df934c-71f0-44da-a52d-730f0a442f68",
	}

	tmpl := template.Must(template.New("domain").Parse(basicDomainXML))

	var domainXML bytes.Buffer
	if err := tmpl.Execute(&domainXML, config); err != nil {
		fmt.Println("oh the error", err)
	}
	d.logger.Info("define libvirt domain using xml: %v", domainXML.String())
	dom, err := d.conn.DomainDefineXML(domainXML.String())
	fmt.Println("oh the error the second time", dom, err)
}
