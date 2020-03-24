package api

/* 
** WS
ENDPOINT: wss://labs.play-with-docker.com/sessions/bprj5cfnctv000el7j5g/ws/
*Doing close session
> {"name": "session close", "args": []}
< {"name": "session end", "args": []}

*Doing init
< {"name":"instance new","args":["bprj727n_bprj77fnctv000el7jn0","192.168.0.48","node1","ip172-18-0-40-bprj727nctv000el7jl0"]}
< {"name":"instance terminal status","args":["bprj727n_bprj77fnctv000el7jn0","connect"]}
< {"name":"instance terminal out","args":["bprj727n_bprj77fnctv000el7jn0","###############################################################\r\n#                          WARNING!!!!                        #\r\n# This is a sandbox environment. Using personal credentials   #\r\n# is HIGHLY! discouraged. Any consequences of doing so are    #\r\n# completely the user's responsibilites.                      #\r\n#                                                             #\r\n# The PWD team.                                               #\r\n###############################################################\r\n"]}
< {"name":"instance terminal out","args":["bprj727n_bprj77fnctv000el7jn0","\u001b[1m\u001b[31m[node1] \u001b[32m(local) \u001b[34mroot@192.168.0.48\u001b[35m ~\u001b[0m\r\r\n$ "]}
< {"name":"instance docker ports","args":[{"instance":"bprj727n_bprj77fnctv000el7jn0","ports":[]}]}
> {"name":"instance viewport resize","args":[135,12]}
< {"name":"instance docker swarm status","args":[{"instance":"bprj727n_bprj77fnctv000el7jn0","is_manager":false,"is_worker":false}]}
< {"name":"instance terminal out","args":["bprj727n_bprj77fnctv000el7jn0","\r\u001b[K$ "]}
< {"name":"instance viewport resize","args":[135,12]}
< {"name":"instance stats","args":[{"cpu":"1.02%","instance":"bprj727n_bprj77fnctv000el7jn0","mem":"0.73% (29.18MiB / 3.906GiB)"}]}
....
> {"name":"instance terminal in","args":["bprj727n_bprj77fnctv000el7jn0","l"]}

< {"name":"instance terminal status","args":["bprjsjnn_bprk0j7nctv000el7nkg","connect"]}
< {"name":"instance terminal out","args":["bprjsjnn_bprk0j7nctv000el7nkg","No such container: bprjsjnn_bprk0j7nctv000el7nkg\r\n"]}

** Endpoints

* Playground Info
> GET https://labs.play-with-docker.com/my/playground
< RES
id: "490c0c31-2304-54c0-ac38-97bf80b045b0"
domain: "labs.play-with-docker.com"
default_dind_instance_image: "franela/dind"
available_dind_instance_images: ["franela/dind", "franela/dind:rc", "franela/go"]
allow_windows_instances: false
default_session_duration: 14400000000000
dind_volume_size: ""

* Image
> GET https://labs.play-with-docker.com/instances/images
< franela/dind, franela/dind:rc, franela/go

* Session Info
> GET https://labs.play-with-docker.com/sessions/bprj5cfnctv000el7j5g
< RES
id: "bprjsjnnctv000el7mtg"
created_at: "2020-03-22T10:30:38.105Z"
expires_at: "2020-03-22T14:30:38.105Z"
pwd_ip_address: "172.18.0.1"
ready: true
stack: ""
stack_name: "pwd"
image_name: ""
host: "40.117.153.174"
user_id: "bpguse7nctv000eom6i0"
playground_id: "490c0c31-2304-54c0-ac38-97bf80b045b0"
instances: {,…}
  bprjsjnn_bprjsvfnctv000el7n10: {name: "bprjsjnn_bprjsvfnctv000el7n10", image: "franela/dind", hostname: "node1", ip: "192.168.0.18",…}
  bprjsjnn_bprjt2vnctv000el7n30: {name: "bprjsjnn_bprjt2vnctv000el7n30", image: "franela/dind", hostname: "node2", ip: "192.168.0.17",…}

* New Instance
> REQ
POST https://labs.play-with-docker.com/sessions/bprhvsfnctv000ftgi1g/instances
{"ImageName":"franela/dind","type":"linux"}
< RES
name: "bprhvsfn_bprj2b7nctv000el7in0"
image: "franela/dind"
hostname: "node3"
ip: "192.168.0.16"
routable_ip: "172.18.0.8"
server_cert: null
server_key: null
ca_cert: null
cert: null
key: null
tls: false
session_id: "bprhvsfnctv000ftgi1g"
proxy_host: "ip172-18-0-8-bprhvsfnctv000ftgi1g"
session_host: "40.117.153.174"
type: ""


* Destroy instance
> REQ
DELETE https://labs.play-with-docker.com/sessions/bprhvsfnctv000ftgi1g/instances/bprhvsfn_bprj2b7nctv000el7in0
< RES
StatusCode: 200

*/

import (
	"os"
	"os/signal"
	"fmt"
	"log"
	"errors"
	"net/url"
	"net/http"
	"bytes"
	"time"
	"strings"
	"encoding/json"

	"github.com/gorilla/websocket"
)

var _ = log.Printf

var (
	ErrSessionNotFound   = errors.New("session is not found")
	ErrInternalError     = errors.New("internal error")
	ErrNotImplemented    = errors.New("not implemented")
	ErrMismatchRCode     = errors.New("mismatch http code")
	ErrSessionIdNotSet   = errors.New("session id needed")
	ErrInstanceNotFound  = errors.New("instance not found")
)

var (
	ENDPOINT_BASE    = "https://%s" // base
	ENDPOINT_SYSTEM  = "/my/playground"
	ENDPOINT_IMAGE   = "/instances/image"
	ENDPOINT_SESSION = "/sessions/%s" // GET [SID]
	ENDPOINT_NEW     = "/sessions/%s/instances" // POST [SID]
	ENDPOINT_DELETE  = "/sessions/%s/instances/%s" // DELETE [SID,INSTANCE_ID]
	ENDPOINT_WS      = "wss://%s/sessions/%s/ws/" // WSS [HOST, SID]
)

/* WS Event */
type evDockerPorts struct {
	Instance string `json:"instance,omitempty"`
	Ports    []int  `json:"ports,omitempty"`
}
type evDockerSwarm struct {
	Instance   string `json:"instance,omitempty"`
	IsManager  bool   `json:"is_manager,omitempty"`
	IsWorker   bool   `json:"is_worker,omitempty"`
}
type evStatus struct {
	Instance string `json:"instance,omitempty"`
	Cpu      string `json:"cpu,omitempty"`
	Memory   string `json:"mem,omitempty"`
}
type wsEvent struct {
	Name string          `json:"name,omitempty"`
	Args []interface{}   `json:"args,omitempty"`
}
func (w wsEvent) GetArrayString() (r []string) {
	for _, i := range w.Args {
		r = append(r, i.(string))
	}
	return r
}
func (w wsEvent) GetArrayInt() (r []int) {
	for _, i := range w.Args {
		r = append(r, i.(int))
	}
	return r
}
///


type instance struct {
	Name        string `json:"name,omitempty"`
	Image       string `json:"image,omitempty"`
	Hostname    string `json:"hostname,omitempty"`
	IP          string `json:"ip,omitempty"`
	RoutableIP  string `json:"routable_ip,omitempty"`
	ServerCert  string `json:"server_cert,omitempty"`
	ServerKey   string `json:"server_key,omitempty"`
	CACert      string `json:"ca_cert,omitempty"`
	Cert        string `json:"cert,omitempty"`
	Key         string `json:"key,omitempty"`
	Tls         bool   `json:"tls,omitempty"`
	SessionId   string `json:"session_id,omitempty"`
	ProxyHost   string `json:"proxy_host,omitempty"`
	SessionHost string `json:"session_host,omitempty"`
	Type        string `json:"type,omitempty"`
}
type playgroundInfo struct {
	ID                          string   `json:"id,omitempty"`
	Domain                      string   `json:"domain,omitempty"`
	DefaultDINDInstanceImage    string   `json:"default_dind_instance_image,omitempty"`
	AvailableDINDInstanceImages []string `json:"available_dind_instance_images,omitempty"`
	AllowWindowsInstances       bool     `json:"allow_windows_instances,omitempty"`
	DefaultSessionDuration      int      `json:"default_session_duration,omitempty"`
	DINDVolumeSize              string   `json:"dind_volume_size,omitempty"`
}
type sessionInfo struct {
	ID           string `json:"id,omitempty"`
	CreateAt     string `json:"created_at,omitempty"`
	ExpiresAt    string `json:"expires_at,omitempty"`
	PWDIPAddress string `json:"pwd_ip_address,omitempty"`
	Ready        bool   `json:"ready,omitempty"`
	Stack        string `json:"stack,omitempty"`
	StackName    string `json:"stack_name,omitempty"`
	ImageName    string `json:"image_name,omitempty"`
	Host         string `json:"host,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	PlaygroundID string `json:"playground_id,omitempty"`
	Instances    map[string]instance `json:"instances,omitempty"`
}
type dashboard struct {
	pwd_url    string
	session_id string
	url        *url.URL
	system     *playgroundInfo
	session    *sessionInfo
}
// req payload
type instanceReq struct {
	ImageName      string
	Type           string
	Hostname       string
	ServerCert     []byte
	ServerKey      []byte
	CACert         []byte
	Cert           []byte
	Key            []byte
	Tls            bool
	PlaygroundFQDN string
	DindVolumeSize string
}

func NewDashboard(pwd_url string) (*dashboard, error) {
	d := &dashboard{
		pwd_url: pwd_url,
	}
	parsed, err := url.Parse(pwd_url)
	if err != nil {
		return nil, err
	}
	d.url = parsed
	d.session_id = strings.TrimPrefix(parsed.Path, "/p/")
	if len(d.session_id) == 0 {
		return nil, fmt.Errorf("%s: must fill", ErrSessionIdNotSet)
	}
	err = d.FetchInfoPg()
	if err != nil {
		return nil, err
	}
	err = d.FetchSessionInfo()
	if err != nil {
		return nil, err
	}
	return d, nil
}
func (d *dashboard) Log(s string, f ...interface{}) {
	///log.Printf("[D]: "+s, f...)
}
func (d *dashboard) FetchInfoPg() error {
	url := fmt.Sprintf(ENDPOINT_BASE+ENDPOINT_SYSTEM, d.url.Host)
	d.Log("url: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %d", ErrMismatchRCode, resp.StatusCode)
	}
	d.system = new(playgroundInfo)
	err = json.NewDecoder(resp.Body).Decode(d.system)
	return err
}
func (d *dashboard) FetchSessionInfo() error {
	url := fmt.Sprintf(ENDPOINT_BASE+ENDPOINT_SESSION, d.url.Host, d.session_id)
	d.Log("url: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ErrSessionNotFound
	}else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %d", ErrMismatchRCode, resp.StatusCode)
	}
	d.session = new(sessionInfo)
	err = json.NewDecoder(resp.Body).Decode(d.session)
	return err
}
func (d *dashboard) GetInstances() (c map[string]instance) {
	if d.session != nil {
		c = map[string]instance{}
		for _,v := range d.session.Instances {
			c[v.Name[12:]] = v
		}
	}
	return c
}

func (d *dashboard) FetchOnWebSocket() error {
	url := fmt.Sprintf(ENDPOINT_WS, d.url.Host, d.session_id)
	fmt.Println(url)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	defer c.Close()
	fmt.Println("connected")
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				d.Log("err: %s", err)
				return
			}
			d.Log("recv: %s", message)
		}
	}()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(`{"name":"instance terminal in","args":["bpt3blac_bpt5s1ac687g00ei7vfg","iiiiiiiiiiiiiiiiiiiiiiiiiiii"]}`))
			fmt.Println("err", err)
		case <-done:
			return nil
		case <-interrupt:
			d.Log("interrupted")
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return nil
		}
	}
	return nil
}

func (d *dashboard) CreateInstanceDefault() (*instance, error) {
	return d.CreateInstance(&instanceReq{
		ImageName: d.system.DefaultDINDInstanceImage,
		Type: "linux",
		DindVolumeSize: "5G", // follow official repo
	})
}
func (d *dashboard) CreateInstance(settings *instanceReq) (*instance, error) {
	url := fmt.Sprintf(ENDPOINT_BASE+ENDPOINT_NEW, d.url.Host, d.session_id)
	d.Log("url: %s\n", url)
	r, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrSessionNotFound
	}else if resp.StatusCode == http.StatusInternalServerError {
		return nil, ErrInternalError
	}else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %d", ErrMismatchRCode, resp.StatusCode)
	}

	i := &instance{}
	err = json.NewDecoder(resp.Body).Decode(i)
	if err != nil {
		return nil, err
	}
	i.Name = i.Name[12:]
	return i, nil
}
func (d *dashboard) DeleteInstance(instanceId string) error {
	url := fmt.Sprintf(ENDPOINT_BASE+ENDPOINT_DELETE, d.url.Host, d.session_id, instanceId)
	d.Log("url: %s", url)
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%s or %s", ErrSessionNotFound, ErrInstanceNotFound)
	}else if resp.StatusCode == http.StatusInternalServerError {
		return ErrInternalError
	}else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %d", ErrMismatchRCode, resp.StatusCode)
	}
	return nil
}