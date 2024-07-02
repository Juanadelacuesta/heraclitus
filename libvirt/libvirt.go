package libvirt

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

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

func (d *driver) getXMLfromConfig(dc *DomainConfig) (string, error) {
	var outb, errb bytes.Buffer

	args := d.parceVirtInstallArgs(dc)

	cmd := exec.Command("virt-install", args...)
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, errb.String())
	}

	return outb.String(), nil
}

func (d *driver) CreateDomain(config *DomainConfig) error {
	domainXML, err := d.getXMLfromConfig(config)
	if err != nil {
		return fmt.Errorf("invalid domain configuration: %w", err)
	}

	//d.logger.Debug("define libvirt domain using xml: %s", domainXML)

	dom, err := d.conn.DomainDefineXML(domainXML)
	if err != nil {
		return fmt.Errorf("unable to define domain: %w", err)
	}

	return dom.Create()
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
		nam, err := dom.GetInfo()
		if err == nil {
			fmt.Printf("  %+v\n", nam)
		}
		dom.Free()
	}
}
