package libvirt

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hashicorp/go-hclog"
	"libvirt.org/go/libvirt"
)

const (
	domainUserDataFolder = "/virt"
	userDataTemplate     = "/libvirt/user-data.tmpl"
	metaDataTemplate     = "/libvirt/meta-data.tmpl"
	envFile              = "/etc/profile.d/virt-envs.sh"
)

var (
	ErrEmptyURI = errors.New("connection URI can't be empty")
)

type driver struct {
	uri     string
	conn    *libvirt.Connect
	logger  hclog.Logger
	dataDir string
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

	return nil
}

func New(ctx context.Context, dataDir string, URI string, logger hclog.Logger) (*driver, error) {
	if err := validURI(URI); err != nil {
		return nil, err
	}

	path := filepath.Join(dataDir, domainUserDataFolder)
	err := os.MkdirAll(path, 0600)
	if err != nil {
		return nil, err
	}

	conn, err := libvirt.NewConnect(URI)
	if err != nil {
		return nil, err
	}

	d := &driver{
		conn:    conn,
		logger:  logger,
		uri:     URI,
		dataDir: path,
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

func (d *driver) createDomain(dc *DomainConfig, ci *cloudinitConfig) error {
	var outb, errb bytes.Buffer

	args := d.parceVirtInstallArgs(dc, ci)

	cmd := exec.Command("virt-install", args...)
	cmd.Dir = d.dataDir
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("libvirt: %w: %s", err, errb.String())
	}
	fmt.Println(outb.String())

	return nil
}

func executeTemplate(config *DomainConfig, in string, out string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("libvirt: unable to get path: %w", err)
	}

	tmpl, err := template.ParseFiles(pwd + in)
	if err != nil {
		return fmt.Errorf("libvirt: unable to parse template: %w", err)
	}

	f, err := os.Create(out)
	defer f.Close()

	if err != nil {
		return fmt.Errorf("libvirt: create file: %w", err)
	}

	err = tmpl.Execute(f, config)
	if err != nil {
		return fmt.Errorf("libvirt: unable to execute template: %w", err)
	}
	return nil
}

func createCloudInitFilesFromTmpls(config *DomainConfig, domainFolder string) (*cloudinitConfig, error) {

	err := executeTemplate(config, metaDataTemplate, domainFolder+"/meta-data")
	if err != nil {
		return nil, err
	}

	err = executeTemplate(config, userDataTemplate, domainFolder+"/user-data")
	if err != nil {
		return nil, err
	}

	ci := &cloudinitConfig{
		userDataPath: domainFolder + "/user-data",
		metadataPath: domainFolder + "/meta-data",
	}

	return ci, nil
}

func (d *driver) CreateDomain(config *DomainConfig) (*libvirt.Domain, error) {
	domainFolder := filepath.Join(d.dataDir, config.Name)
	err := os.MkdirAll(domainFolder, 0700)
	if err != nil {
		return nil, err
	}

	ci, err := createCloudInitFilesFromTmpls(config, domainFolder)
	if err != nil {
		return nil, err
	}

	err = d.createDomain(config, ci)
	if err != nil {
		return nil, err
	}
	dom, err := d.conn.LookupDomainByName(config.Name)
	fmt.Println("invisible error?", err)
	j, err := dom.GetID()
	fmt.Println("getting the ID maybe: ", j, err)

	return dom, err
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
