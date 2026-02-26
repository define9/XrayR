package sparrowpanel

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/go-resty/resty/v2"
	"github.com/xtls/xray-core/infra/conf"

	"github.com/XrayR-project/XrayR/api"
)

// APIClient create an api client to the panel.
type APIClient struct {
	client        *resty.Client
	APIHost       string
	NodeID        int
	Key           string
	NodeType      string
	EnableVless   bool
	VlessFlow     string
	SpeedLimit    float64
	DeviceLimit   int
	LocalRuleList []api.DetectRule
}

// New creat an api instance
func New(apiConfig *api.Config) *APIClient {
	client := resty.New()
	client.SetRetryCount(3)
	if apiConfig.Timeout > 0 {
		client.SetTimeout(time.Duration(apiConfig.Timeout) * time.Second)
	} else {
		client.SetTimeout(5 * time.Second)
	}
	client.OnError(func(req *resty.Request, err error) {
		var v *resty.ResponseError
		if errors.As(err, &v) {
			// v.Response contains the last response from the server
			// v.Err contains the original error
			log.Print(v.Err)
		}
	})
	client.SetBaseURL(apiConfig.APIHost)
	// Create Key for each requests
	client.SetHeaders(map[string]string{
		"SparrowPanelApiKey": apiConfig.Key, //apiConfig.Key,
	})
	// Read local rule list
	localRuleList := readLocalRuleList(apiConfig.RuleListPath)
	apiClient := &APIClient{
		client:        client,
		NodeID:        apiConfig.NodeID,
		Key:           apiConfig.Key,
		APIHost:       apiConfig.APIHost,
		NodeType:      apiConfig.NodeType,
		EnableVless:   apiConfig.EnableVless,
		VlessFlow:     apiConfig.VlessFlow,
		SpeedLimit:    apiConfig.SpeedLimit,
		DeviceLimit:   apiConfig.DeviceLimit,
		LocalRuleList: localRuleList,
	}
	return apiClient
}

// readLocalRuleList reads the local rule list file
func readLocalRuleList(path string) (LocalRuleList []api.DetectRule) {
	LocalRuleList = make([]api.DetectRule, 0)
	if path != "" {
		// open the file
		file, err := os.Open(path)

		// handle errors while opening
		if err != nil {
			log.Printf("Error when opening file: %s", err)
			return LocalRuleList
		}

		fileScanner := bufio.NewScanner(file)

		// read line by line
		for fileScanner.Scan() {
			LocalRuleList = append(LocalRuleList, api.DetectRule{
				ID:      -1,
				Pattern: regexp.MustCompile(fileScanner.Text()),
			})
		}
		// handle first encountered error while reading
		if err := fileScanner.Err(); err != nil {
			log.Fatalf("Error while reading file: %s", err)
			return
		}

		_ = file.Close()
	}

	return LocalRuleList
}

// Describe return a description of the client
func (c *APIClient) Describe() api.ClientInfo {
	return api.ClientInfo{APIHost: c.APIHost, NodeID: c.NodeID, Key: c.Key, NodeType: c.NodeType}
}

// Debug set the client debug for client
func (c *APIClient) Debug() {
	c.client.SetDebug(true)
}

func (c *APIClient) assembleURL(path string) string {
	return c.APIHost + path
}

func (c *APIClient) parseResponse(res *resty.Response, path string, err error) (*Response, error) {
	if err != nil {
		return nil, fmt.Errorf("request %s failed: %s", c.assembleURL(path), err)
	}

	if res.StatusCode() > 400 {
		body := res.Body()
		return nil, fmt.Errorf("request %s failed: %s, %s", c.assembleURL(path), string(body), err)
	}
	response := res.Result().(*Response)

	if response.Code != "000-000" {
		res, _ := json.Marshal(&response)
		log.Printf("res error, message: %s", response.Message)
		return nil, fmt.Errorf("ret %s invalid", string(res))
	}
	return response, nil
}

// GetNodeInfo will pull NodeInfo Config from SparrowPanel
func (c *APIClient) GetNodeInfo() (nodeInfo *api.NodeInfo, err error) {
	path := fmt.Sprintf("/api/v1/proxyServer/xray/nodeInfo")

	res, err := c.client.R().
		SetQueryParams(map[string]string{
			"type":   c.NodeType,
			"nodeId": strconv.Itoa(c.NodeID),
		}).
		SetResult(&Response{}).
		ForceContentType("application/json").
		Get(path)

	response, err := c.parseResponse(res, path, err)
	if err != nil {
		return nil, err
	}

	nodeInfoResponse := new(NodeInfo)

	if err := json.Unmarshal(response.Data, nodeInfoResponse); err != nil {
		return nil, fmt.Errorf("unmarshal %s failed: %s", reflect.TypeOf(nodeInfoResponse), err)
	}
	switch c.NodeType {
	case "V2ray":
		nodeInfo, err = c.ParseV2rayNodeResponse(nodeInfoResponse)
	case "Trojan":
		nodeInfo, err = c.ParseTrojanNodeResponse(nodeInfoResponse)
	case "Shadowsocks":
		nodeInfo, err = c.ParseSSNodeResponse(nodeInfoResponse)
	default:
		return nil, fmt.Errorf("unsupported Node type: %s", c.NodeType)
	}

	if err != nil {
		res, _ := json.Marshal(nodeInfoResponse)
		return nil, fmt.Errorf("Parse node info failed: %s, \nError: %s", string(res), err)
	}

	return nodeInfo, nil
}

// GetUserList will pull user form sparrow panel
func (c *APIClient) GetUserList() (UserList *[]api.UserInfo, err error) {
	path := "/api/v1/proxyServer/xray/users"
	res, err := c.client.R().
		SetQueryParams(map[string]string{
			"type":   c.NodeType,
			"nodeId": strconv.Itoa(c.NodeID),
		}).
		SetResult(&Response{}).
		ForceContentType("application/json").
		Get(path)

	response, err := c.parseResponse(res, path, err)
	if err != nil {
		return nil, err
	}

	var userListResponse *[]User
	if err := json.Unmarshal(response.Data, &userListResponse); err != nil {
		return nil, fmt.Errorf("unmarshal %s failed: %s", reflect.TypeOf(userListResponse), err)
	}
	userList, err := c.ParseUserListResponse(userListResponse)
	if err != nil {
		res, _ := json.Marshal(userListResponse)
		return nil, fmt.Errorf("parse user list failed: %s", string(res))
	}
	return userList, nil
}

// ReportNodeStatus reports the node status to the sparrow panel
func (c *APIClient) ReportNodeStatus(nodeStatus *api.NodeStatus) (err error) {
	status := NodeStatus{
		Uptime: nodeStatus.Uptime,
		CPU:    nodeStatus.CPU,
		Mem:    nodeStatus.Mem,
		Disk:   nodeStatus.Disk,
	}

	path := "/api/v1/proxyServer/xray/reportNodeStatus"
	res, err := c.client.R().
		SetQueryParams(map[string]string{
			"type":   c.NodeType,
			"nodeId": strconv.Itoa(c.NodeID),
		}).
		SetBody(status).
		SetResult(&Response{}).
		ForceContentType("application/json").
		Post(path)

	_, err = c.parseResponse(res, path, err)
	if err != nil {
		return err
	}

	return nil
}

// ReportNodeOnlineUsers reports online user ip
func (c *APIClient) ReportNodeOnlineUsers(onlineUserList *[]api.OnlineUser) error {
	data := make([]OnlineUser, len(*onlineUserList))
	for i, user := range *onlineUserList {
		data[i] = OnlineUser{UID: user.UID, IP: user.IP}
	}
	postData := &PostData{NodeType: c.NodeType, NodeId: c.NodeID, Onlines: data}
	path := "/api/v1/proxyServer/xray/online"

	res, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(postData).
		SetResult(&Response{}).
		ForceContentType("application/json").
		Post(path)
	_, err = c.parseResponse(res, path, err)
	if err != nil {
		return err
	}

	return nil
}

// ReportUserTraffic reports the user traffic
func (c *APIClient) ReportUserTraffic(userTraffic *[]api.UserTraffic) error {
	data := make([]UserTraffic, len(*userTraffic))
	for i, traffic := range *userTraffic {
		data[i] = UserTraffic{
			UID:      traffic.UID,
			Upload:   traffic.Upload,
			Download: traffic.Download,
		}
	}
	postData := &PostData{NodeType: c.NodeType, NodeId: c.NodeID, Users: data}
	path := "/api/v1/proxyServer/xray/traffic"

	res, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(postData).
		SetResult(&Response{}).
		ForceContentType("application/json").
		Post(path)
	_, err = c.parseResponse(res, path, err)
	if err != nil {
		return err
	}

	return nil
}

// GetNodeRule will pull the audit rule form panel
func (c *APIClient) GetNodeRule() (*[]api.DetectRule, error) {
	ruleList := c.LocalRuleList
	path := "/api/v1/proxyServer/xray/rules"

	res, err := c.client.R().
		SetQueryParams(map[string]string{
			"type":   c.NodeType,
			"nodeId": strconv.Itoa(c.NodeID),
		}).
		SetResult(&Response{}).
		ForceContentType("application/json").
		Get(path)

	response, err := c.parseResponse(res, path, err)
	if err != nil {
		return nil, err
	}

	ruleListResponse := new([]RuleItem)

	if err := json.Unmarshal(response.Data, ruleListResponse); err != nil {
		return nil, fmt.Errorf("unmarshal %s failed: %s", reflect.TypeOf(ruleListResponse), err)
	}

	for _, r := range *ruleListResponse {
		ruleList = append(ruleList, api.DetectRule{
			ID:      r.ID,
			Pattern: regexp.MustCompile(r.Content),
		})
	}
	return &ruleList, nil
}

// ReportIllegal reports the user illegal behaviors
func (c *APIClient) ReportIllegal(detectResultList *[]api.DetectResult) error {
	return nil
}

// ParseV2rayNodeResponse parse the response for the given nodeInfo format
func (c *APIClient) ParseV2rayNodeResponse(s *NodeInfo) (*api.NodeInfo, error) {
	var (
		host      string
		header    json.RawMessage
		enableTLS bool
	)

	switch s.Network {
	case "WebSocket":
		if s.NetworkSettings.Headers != nil {
			if httpHeader, err := s.NetworkSettings.Headers.MarshalJSON(); err != nil {
				return nil, err
			} else {
				b, _ := simplejson.NewJson(httpHeader)
				host = b.Get("Host").MustString()
			}
		}
	case "TCP":
		if s.NetworkSettings.Header != nil {
			if httpHeader, err := s.NetworkSettings.Header.MarshalJSON(); err != nil {
				return nil, err
			} else {
				header = httpHeader
			}
		}
	case "GRPC":
	}

	enableTLS = s.EnableTLS

	// Create GeneralNodeInfo
	return &api.NodeInfo{
		NodeType:          c.NodeType,
		NodeID:            c.NodeID,
		Port:              uint32(s.ServerPort),
		AlterID:           0,
		TransportProtocol: s.Network,
		EnableTLS:         enableTLS,
		Path:              s.NetworkSettings.Path,
		Host:              host,
		EnableVless:       c.EnableVless,
		VlessFlow:         c.VlessFlow,
		ServiceName:       s.NetworkSettings.ServiceName,
		Header:            header,
		NameServerConfig:  s.parseDNSConfig(),
	}, nil
}

// ParseSSNodeResponse parse the response for the given nodeInfo format
func (c *APIClient) ParseSSNodeResponse(nodeInfoResponse *NodeInfo) (*api.NodeInfo, error) {

	// Create GeneralNodeInfo
	nodeInfo := &api.NodeInfo{}

	return nodeInfo, nil
}

// ParseTrojanNodeResponse parse the response for the given nodeInfo format
func (c *APIClient) ParseTrojanNodeResponse(nodeInfoResponse *NodeInfo) (*api.NodeInfo, error) {

	// Create GeneralNodeInfo
	nodeInfo := &api.NodeInfo{}

	return nodeInfo, nil
}

// ParseUserListResponse parse the response for the given nodeInfo format
func (c *APIClient) ParseUserListResponse(userInfoResponse *[]User) (*[]api.UserInfo, error) {
	var deviceLimit = 0
	var speedLimit uint64 = 0
	userList := make([]api.UserInfo, len(*userInfoResponse))
	for i, user := range *userInfoResponse {
		if c.DeviceLimit > 0 {
			deviceLimit = c.DeviceLimit
		} else {
			deviceLimit = 3
		}
		if c.SpeedLimit > 0 {
			speedLimit = uint64((c.SpeedLimit * 1000000) / 8)
		} else {
			speedLimit = uint64((user.SpeedLimit * 1000000) / 8)
		}
		userList[i] = api.UserInfo{
			UID:         user.Id,
			Passwd:      user.NodePassword,
			UUID:        user.NodePassword,
			SpeedLimit:  speedLimit,
			DeviceLimit: deviceLimit,
		}
	}

	return &userList, nil
}

func (s *NodeInfo) parseDNSConfig() (nameServerList []*conf.NameServerConfig) {
	return
}
