package vpsie

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client interface {
	GetBalance() (Balance, error)
	GetProcessStatus(processId string) (ProcessStatus, error)

	GetOffers() ([]Offer, error)
	GetDatacenters() ([]Datacenter, error)
	GetImages() ([]Image, error)

	CreateVPSie(create CreateVPSie) (VPSie, error)
	DeleteVPSie(id string) (string, error)
	GetVPSie(id string) (VPSie, error)
	ListVPSie() ([]VPSie, error)
	StartVPSie(id string) (string, error)
	ShutdownVPSie(id string) (VPSieActionResponse, error)
	RestartVPSie(id string) (string, error)
	ForceRestartVPSie(id string) (string, error)
	ChangeVPSieHostname(id string, hostname string) (VPSieActionResponse, error)
	ChangeVPSiePassword(id string) (VPSiePasswordResponse, error)
	BackupVPSie(id string, name string, note string) (VPSieBackupResponse, error)
	SnapshotVPSie(id string, name string, note string) (VPSieSnapshotResponse, error)
	ResizeVPSie(id string, cpu string, ssd string, ram string) (VPSieActionResponse, error)
	RebuildVPSie(id string) (VPSieRebuildResponse, error)
	VPSieStatistics(id string) (VPSieStatisticsResponse, error)
}

type client struct {
	clientId     string
	clientSecret string
	accessToken  string
	expires      time.Time
	refreshToken string
	debug        bool
}

func NewClient(clientId string, clientSecret string, debug bool) Client {
	return &client{clientId: clientId, clientSecret: clientSecret, debug: debug}
}

func (c *client) doGetRequest(action string, out interface{}) error {
	req, err := http.NewRequest("GET", "https://api.vpsie.com/v1/"+action, nil)
	if err != nil {
		return err
	}

	if bearerToken, err := c.getBearerToken(); err != nil {
		return err
	} else {
		req.Header.Add("Authorization", "Bearer "+bearerToken)
	}

	if c.debug {
		log.Println("GET", action)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if c.debug {
		log.Println("<===", string(content))
	}

	return json.Unmarshal(content, &out)
}

func (c *client) doPostRequest(action string, in url.Values, out interface{}) error {
	encodedForm := in.Encode()

	req, err := http.NewRequest("POST", "https://api.vpsie.com/v1/"+action, strings.NewReader(encodedForm))
	if err != nil {
		return err
	}

	if bearerToken, err := c.getBearerToken(); err != nil {
		return err
	} else {
		req.Header.Add("Authorization", "Bearer "+bearerToken)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if c.debug {
		log.Println("POST", action)
		log.Println("===>", encodedForm)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if c.debug {
		log.Println("<===", string(content))
	}

	return json.Unmarshal(content, &out)
}

func (c *client) doDeleteRequest(action string, out interface{}) error {
	req, err := http.NewRequest("DELETE", "https://api.vpsie.com/v1/"+action, nil)
	if err != nil {
		return err
	}

	if bearerToken, err := c.getBearerToken(); err != nil {
		return err
	} else {
		req.Header.Add("Authorization", "Bearer "+bearerToken)
	}

	if c.debug {
		log.Println("DELETE", action)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if c.debug {
		log.Println("<===", string(content))
	}

	return json.Unmarshal(content, &out)
}

type baseResponse struct {
	Error     bool   `json:"error"`
	ErrorCode string `json:"errorCode"`
}

type token struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

type authenticateResponse struct {
	baseResponse
	Token token `json:"token"`
}

func (c *client) getBearerToken() (string, error) {
	if c.accessToken != "" && time.Now().UTC().Before(c.expires) {
		return c.accessToken, nil
	}

	data := url.Values{}
	data.Add("grand_type", "bearer")
	data.Add("client_id", c.clientId)
	data.Add("client_secret", c.clientSecret)
	res, err := http.PostForm("https://api.vpsie.com/v1/token", data)
	if err != nil {
		return "", err
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	authRes := authenticateResponse{}
	if err := json.Unmarshal(bytes, &authRes); err != nil {
		return "", err
	}

	c.accessToken = authRes.Token.AccessToken
	c.refreshToken = authRes.Token.RefreshToken
	c.expires = time.Now().UTC().Add(time.Duration(authRes.Token.ExpiresIn))

	return c.accessToken, nil
}
