package driver

import (
	"errors"
	"fmt"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/jdextraze/go-vpsie"
	"io/ioutil"
)

const (
	defaultOfferID      = "9a0e49c6-9f22-11e3-8af5-005056aa8af7"
	defaultDatacenterID = "55f06b85-c9ee-11e3-9845-005056aa8af7"
	defaultImageID      = "75401d7d-d9d3-11e3-b135-005056aa8af7"
	SSHUser             = "root"
	SSHPort             = 22
)

type Driver struct {
	*drivers.BaseDriver
	ClientId     string
	ClientSecret string

	ImageID      string
	OfferID      string
	DatacenterID string

	InstanceID string

	client vpsie.Client
}

func NewDriver(hostName, storePath string) *Driver {
	d := &Driver{
		ImageID:      defaultImageID,
		OfferID:      defaultOfferID,
		DatacenterID: defaultDatacenterID,
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
			SSHUser:     SSHUser,
			SSHPort:     SSHPort,
		},
	}
	return d
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "VPSIE_CLIENT_ID",
			Name:   "vpsie-client-id",
			Usage:  "VPSie Client ID",
		},
		mcnflag.StringFlag{
			EnvVar: "VPSIE_CLIENT_SECRET",
			Name:   "vpsie-client-secret",
			Usage:  "VPSie Client secret",
		},
		mcnflag.StringFlag{
			EnvVar: "VPSIE_IMAGE_ID",
			Name:   "vpsie-image-id",
			Usage:  "VPSie Image ID",
			Value:  defaultImageID,
		},
		mcnflag.StringFlag{
			EnvVar: "VPSIE_OFFER_ID",
			Name:   "vpsie-offer-id",
			Usage:  "VPSie Offer ID",
			Value:  defaultOfferID,
		},
		mcnflag.StringFlag{
			EnvVar: "VPSIE_DATACENTER_ID",
			Name:   "vpsie-datacenter-id",
			Usage:  "VPSie Datacenter ID",
			Value:  defaultDatacenterID,
		},
	}
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) DriverName() string {
	return "vpsie"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.ClientId = flags.String("vpsie-client-id")
	d.ClientSecret = flags.String("vpsie-client-secret")
	d.ImageID = flags.String("vpsie-image-id")
	d.DatacenterID = flags.String("vpsie-datacenter-id")
	d.OfferID = flags.String("vpsie-offer-id")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")

	if d.ClientId == "" {
		return fmt.Errorf("VPSie driver requires the --vpsie-client-id option")
	}
	if d.ClientSecret == "" {
		return fmt.Errorf("VPSie driver requires the --vpsie-client-secret option")
	}
	return nil
}

func (d *Driver) PreCreateCheck() error {
	log.Info("Validating VPSie VPS parameters...")

	if err := d.validateImageID(); err != nil {
		return err
	}

	if err := d.validateDatacenterID(); err != nil {
		return err
	}

	if err := d.validateOfferID(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Create() error {
	log.Info("Creating VPSie VPS...")

	sshKey, err := d.createSSHKey()
	if err != nil {
		return err
	}

	instance, err := d.getClient().CreateVPSie(vpsie.CreateVPSie{
		Hostname:     d.MachineName,
		OfferId:      d.OfferID,
		DatacenterId: d.DatacenterID,
		OsId:         d.ImageID,
	})
	if err != nil {
		return err
	}
	d.InstanceID = instance.Id
	d.IPAddress = instance.IpV4

	log.Infof("Created VPSie VPS ID: %s, Public IP: %s",
		d.InstanceID,
		d.IPAddress,
	)

	d.addSshKeyToServer(instance.Password, sshKey)

	return nil
}

func (d *Driver) GetURL() (string, error) {
	s, err := d.GetState()
	if err != nil {
		return "", err
	}

	if s != state.Running {
		return "", drivers.ErrHostIsNotRunning
	}

	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	if d.IPAddress == "" || d.IPAddress == "0" {
		return "", fmt.Errorf("IP address is not set")
	}
	return d.IPAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	machine, err := d.getClient().GetVPSie(d.InstanceID)
	if err != nil {
		return state.Error, err
	}
	switch machine.Status {
	case "Started":
		return state.Starting, nil
	case "Running":
		return state.Running, nil
	case "Stopped":
		return state.Stopped, nil
	}
	return state.Error, nil
}

func (d *Driver) Start() error {
	status, err := d.getClient().StartVPSie(d.InstanceID)
	if err != nil {
		return err
	} else if status != "Started" {
		return fmt.Errorf("Invalid status %s after start", status)
	}
	return nil
}

func (d *Driver) Stop() error {
	actionStatus, err := d.getClient().ShutdownVPSie(d.InstanceID)
	if err != nil {
		return err
	} else if actionStatus.Error {
		return errors.New(actionStatus.ErrorCode)
	}
	return nil
}

func (d *Driver) Remove() error {
	status, err := d.getClient().DeleteVPSie(d.InstanceID)
	if err != nil {
		return err
	} else if status != "Deleted" {
		return fmt.Errorf("Invalid status %s after remove", status)
	}
	return nil
}

func (d *Driver) Restart() error {
	status, err := d.getClient().RestartVPSie(d.InstanceID)
	if err != nil {
		return err
	} else if status != "Restarted" {
		return fmt.Errorf("Invalid status %s after restart", status)
	}
	return nil
}

func (d *Driver) Kill() error {
	actionStatus, err := d.getClient().ShutdownVPSie(d.InstanceID)
	if err != nil {
		return err
	} else if actionStatus.Error {
		return errors.New(actionStatus.ErrorCode)
	}
	return nil
}

func (d *Driver) getClient() vpsie.Client {
	log.Debug("getting client")
	if d.client == nil {
		d.client = vpsie.NewClient(d.ClientId, d.ClientSecret, true)
	}
	return d.client
}

func (d *Driver) validateImageID() error {
	images, err := d.getClient().GetImages()
	if err != nil {
		return err
	}

	for _, image := range images {
		if image.Id == d.ImageID {
			return nil
		}
	}

	return fmt.Errorf("Image ID %s is invalid", d.ImageID)
}

func (d *Driver) validateDatacenterID() error {
	datacenters, err := d.getClient().GetDatacenters()
	if err != nil {
		return err
	}

	for _, datacenter := range datacenters {
		if datacenter.Id == d.DatacenterID {
			return nil
		}
	}

	return fmt.Errorf("Datacenter ID %s is invalid", d.DatacenterID)
}

func (d *Driver) validateOfferID() error {
	offers, err := d.getClient().GetOffers()
	if err != nil {
		return err
	}

	for _, offer := range offers {
		if offer.Id == d.OfferID {
			return nil
		}
	}

	return fmt.Errorf("Offer ID %s is invalid", d.OfferID)
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) createSSHKey() ([]byte, error) {
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return nil, err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

func (d *Driver) addSshKeyToServer(password string, sshKey []byte) error {
	log.Info("Waiting for machine to be running, this may take a few minutes...")
	if err := mcnutils.WaitFor(drivers.MachineInState(d, state.Running)); err != nil {
		return fmt.Errorf("Error waiting for machine to be running: %s", err)
	}

	log.Info("Waiting for SSH to be available...")
	if err := mcnutils.WaitFor(d.sshAvailableFunc(password)); err != nil {
		return fmt.Errorf("Error waiting for ssh to be available: %s", err)
	}

	_, err := d.runSshCommand(
		password,
		"mkdir ~/.ssh && echo '"+string(sshKey)+"' >> ~/.ssh/authorized_keys",
	)
	return err
}

func (d *Driver) sshAvailableFunc(password string) func() bool {
	return func() bool {
		log.Debug("Getting to WaitForSSH function...")
		if _, err := d.runSshCommand(password, "exit 0"); err != nil {
			log.Debugf("Error getting ssh command 'exit 0' : %s", err)
			return false
		}
		return true
	}
}

func (d *Driver) runSshCommand(password string, cmd string) (string, error) {
	c, err := d.getSshClient(password)
	if err != nil {
		return "", err
	}

	out, err := c.Output(cmd)
	log.Debugf("Ssh command output: %s", out)

	return out, err
}

func (d *Driver) getSshClient(password string) (ssh.Client, error) {
	address, err := d.GetSSHHostname()
	if err != nil {
		return nil, err
	}

	port, err := d.GetSSHPort()
	if err != nil {
		return nil, err
	}

	auth := &ssh.Auth{
		Passwords: []string{password},
	}

	ssh.SetDefaultClient(ssh.Native)

	return ssh.NewClient(d.GetSSHUsername(), address, port, auth)
}
