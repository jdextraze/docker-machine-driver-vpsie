package vpsie

import (
	"net/url"
	"time"
)

type CreateVPSie struct {
	Hostname     string
	OfferId      string
	DatacenterId string
	OsId         string
	IpV6         *bool
	AutoBackup   *bool
	IpV4         *bool
	PrivateIp    *bool
	Note         *string
}

type VPSie struct {
	Id           string    `json:"vpsie_id"`
	Name         string    `json:"name"`
	Ram          int       `json:"ram"`
	Ssd          int       `json:"ssd"`
	Cpu          int       `json:"cpu"`
	Bandwith     int       `json:"bandwith"`
	Note         string    `json:"note"`
	Password     string    `json:"password"`
	OsSlug       string    `json:"os_slug"`
	Distribution string    `json:"distribution"`
	Region       string    `json:"region"`
	IpV4         string    `json:"ipv4"`
	IpV6         string    `json:"ipv6"`
	PrivateIp    string    `json:"private_ip"`
	CreatedOn    time.Time `json:"created_on"`
	Status       string    `json:"status"`
	ProcessId    string    `json:"process_id"`
}

func (c *client) CreateVPSie(create CreateVPSie) (VPSie, error) {
	in := url.Values{}
	in.Add("hostname", create.Hostname)
	in.Add("offer_id", create.OfferId)
	in.Add("datacenter_id", create.DatacenterId)
	in.Add("os_id", create.OsId)
	if create.IpV6 != nil && *create.IpV6 {
		in.Add("ipv6", "true")
	}
	if create.AutoBackup != nil && *create.AutoBackup {
		in.Add("autobackup", "true")
	}
	if create.IpV4 != nil && !*create.IpV4 {
		in.Add("ipv4", "false")
	}
	if create.PrivateIp != nil && *create.PrivateIp {
		in.Add("private_ip", "true")
	}
	if create.Note != nil {
		in.Add("note", *create.Note)
	}

	out := VPSie{}
	return out, c.doPostRequest("vpsie", in, &out)
}

type VPSiesResponse struct {
	baseResponse
	VPSies []VPSie `json:"vpsies"`
}

func (c *client) ListVPSie() ([]VPSie, error) {
	out := VPSiesResponse{}
	return out.VPSies, c.doGetRequest("vpsies", &out)
}

type VPSieStatusResponse struct {
	baseResponse
	Status string `json:"status"`
}

func (c *client) DeleteVPSie(id string) (string, error) {
	out := VPSieStatusResponse{}
	return out.Status, c.doDeleteRequest("vpsie/"+id, &out)
}

func (c *client) GetVPSie(id string) (VPSie, error) {
	out := VPSie{}
	return out, c.doGetRequest("vpsie/"+id, &out)
}

func (c *client) StartVPSie(id string) (string, error) {
	out := VPSieStatusResponse{}
	return out.Status, c.doPostRequest("vpsie/start/"+id, nil, &out)
}

type VPSieActionResponse struct {
	baseResponse
	Action    string `json:"action"`
	Status    string `json:"status"`
	ProcessId string `json:"process_id"`
}

func (c *client) ShutdownVPSie(id string) (VPSieActionResponse, error) {
	out := VPSieActionResponse{}
	return out, c.doPostRequest("vpsie/shutdown/"+id, nil, &out)
}

func (c *client) RestartVPSie(id string) (string, error) {
	out := VPSieStatusResponse{}
	return out.Status, c.doPostRequest("vpsie/restart/"+id, nil, &out)
}

func (c *client) ForceRestartVPSie(id string) (string, error) {
	out := VPSieStatusResponse{}
	return out.Status, c.doPostRequest("vpsie/force/restart/"+id, nil, &out)
}

func (c *client) ChangeVPSieHostname(id string, hostname string) (VPSieActionResponse, error) {
	in := url.Values{}
	in.Add("hostname", hostname)
	out := VPSieActionResponse{}
	return out, c.doPostRequest("vpsie/rename/"+id, in, &out)
}

type VPSiePasswordResponse struct {
	VPSieActionResponse
	Password string `json:"password"`
}

func (c *client) ChangeVPSiePassword(id string) (VPSiePasswordResponse, error) {
	out := VPSiePasswordResponse{}
	return out, c.doPostRequest("vpsie/password/"+id, nil, &out)
}

type VPSieBackupResponse struct {
	VPSieActionResponse
	BackupName string `json:"backupname"`
}

func (c *client) BackupVPSie(id string, name string, note string) (VPSieBackupResponse, error) {
	in := url.Values{}
	in.Add("backupname", name)
	if note != "" {
		in.Add("backupnote", note)
	}
	out := VPSieBackupResponse{}
	return out, c.doPostRequest("vpsie/backup/"+id, in, &out)
}

type VPSieSnapshotResponse struct {
	VPSieActionResponse
	SnaphotName string `json:"snapname"`
}

func (c *client) SnapshotVPSie(id string, name string, note string) (VPSieSnapshotResponse, error) {
	in := url.Values{}
	in.Add("snapname", name)
	if note != "" {
		in.Add("snapnote", note)
	}
	out := VPSieSnapshotResponse{}
	return out, c.doPostRequest("vpsie/snapshot/"+id, in, &out)
}

func (c *client) ResizeVPSie(id string, cpu string, ssd string, ram string) (VPSieActionResponse, error) {
	in := url.Values{}
	in.Add("cpu", cpu)
	in.Add("ssd", ssd)
	in.Add("ram", ram)
	out := VPSieActionResponse{}
	return out, c.doPostRequest("vpsie/resize/"+id, in, &out)
}

type VPSieRebuildResponse struct {
	VPSieActionResponse
	NewVPSieId string `json:"new_vpsie_id"`
}

func (c *client) RebuildVPSie(id string) (VPSieRebuildResponse, error) {
	out := VPSieRebuildResponse{}
	return out, c.doPostRequest("vpsie/rebuild/"+id, nil, &out)
}

type graph struct {
	Cpu       []float32 `json:"cpu"`
	Ram       []float32 `json:"ram"`
	DiskRead  []int64   `json:"diskread"`
	DiskWrite []int64   `json:"diskwrite"`
	NetIn     []int64   `json:"netin"`
	NetOut    []int64   `json:"netout"`
	Time      []string  `json:"time"`
}

type VPSieStatisticsResponse struct {
	baseResponse
	Graph graph `json:"graph"`
}

func (c *client) VPSieStatistics(id string) (VPSieStatisticsResponse, error) {
	out := VPSieStatisticsResponse{}
	return out, c.doPostRequest("vpsie/statistics/"+id, nil, &out)
}
