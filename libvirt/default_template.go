package libvirt

const domainTmpl = `
<domain type='kvm'>
  <name>{{.MachineName}}</name>
  <memory unit='MiB'>{{.Memory}}</memory>
  <vcpu>{{.CPU}}</vcpu>
  <features>
    <acpi/>
    <apic/>
    <pae/>
    {{if .Hidden}}
    <kvm>
      <hidden state='on'/>
    </kvm>
    {{end}}
  </features>
  <cpu mode='host-passthrough'>
  {{if gt .NUMANodeCount 1}}
  {{.NUMANodeXML}}
  {{end}}
  </cpu>
  <os>
    <type>hvm</type>
    <boot dev='cdrom'/>
    <boot dev='hd'/>
    <bootmenu enable='no'/>
  </os>
  <devices>
    <disk type='file' device='cdrom'>
      <source file='{{.ISO}}'/>
      <target dev='hdc' bus='scsi'/>
      <readonly/>
    </disk>
    <disk type='file' device='disk'>
      <driver name='qemu' type='raw' cache='default' io='threads' />
      <source file='{{.DiskPath}}'/>
      <target dev='hda' bus='virtio'/>
    </disk>
    <interface type='network'>
      <source network='{{.PrivateNetwork}}'/>
      <model type='virtio'/>
    </interface>
    <interface type='network'>
      <source network='{{.Network}}'/>
      <model type='virtio'/>
    </interface>
    <serial type='pty'>
      <target port='0'/>
    </serial>
    <console type='pty'>
      <target type='serial' port='0'/>
    </console>
    <rng model='virtio'>
      <backend model='random'>/dev/random</backend>
    </rng>
    {{if .GPU}}
    {{.DevicesXML}}
    {{end}}
    {{if gt .ExtraDisks 0}}
    {{.ExtraDisksXML}}
    {{end}}
  </devices>
</domain>
`
const basicDomainXML = `
<domain type='kvm'>
  <name>{{.Name}}</name>
  <uuid>{{.ID}}</uuid>
  <memory unit='KiB'>524288</memory> <!-- 512 MB of RAM -->
  <vcpu placement='static'>1</vcpu>
  <os>
    <type arch='x86_64' machine='pc-i440fx-2.9'>hvm</type>
    <boot dev='hd'/>
  </os>
  <devices>
    <disk type='file' device='cdrom'>
      <driver name='qemu' type='raw'/>
      <source file='/var/lib/libvirt/images/jammy-server-cloudimg-amd64.img'/>
      <target dev='hdb' bus='ide'/>
      <readonly/>
      <address type='pci' domain='0x0000' bus='0x01' slot='0x01' function='0x0'/>
    </disk>
    <interface type='network'>
      <mac address='52:54:00:6b:3c:58'/>
      <source network='default'/>
      <model type='virtio'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x03' function='0x0'/>
    </interface>
    <console type='pty'>
      <target type='serial' port='0'/>
    </console>
  </devices>
</domain>
`
