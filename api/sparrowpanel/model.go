package sparrowpanel

import (
	"encoding/json"
)

// Response is the common response
type Response struct {
	Code    string          `json:"code"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

// NodeInfo is the response of node
type NodeInfo struct {
	V2ray
	Shadowsocks
	Trojan

	ServerPort int    `json:"serverPort"`
	Port       uint32 `json:"port"`
}

type ServerConfig struct {
	NodeInfo

	BaseConfig struct {
		PushInterval int `json:"push_interval"`
		PullInterval int `json:"pull_interval"`
	} `json:"base_config"`
}

type Shadowsocks struct {
	Cipher       string `json:"cipher"`
	Obfs         string `json:"obfs"`
	ObfsSettings struct {
		Path string `json:"path"`
		Host string `json:"host"`
	} `json:"obfs_settings"`
	ServerKey string `json:"server_key"`
}

type V2ray struct {
	Network         string `json:"network"`
	NetworkSettings struct {
		Path        string           `json:"path"`
		Headers     *json.RawMessage `json:"headers"`
		ServiceName string           `json:"serviceName"`
		Header      *json.RawMessage `json:"header"`
	} `json:"networkSettings"`
	EnableTLS bool `json:"enableTls"`
}

type Trojan struct {
	Host       string `json:"host"`
	ServerName string `json:"name"`
}

type User struct {
	Id           int    `json:"id"`
	NodePassword string `json:"nodePassword"`
	SpeedLimit   int    `json:"speedLimit"`
}

// PostData is the data structure of post data
type PostData struct {
	NodeType string      `json:"nodeType"`
	NodeId   int         `json:"nodeId"`
	Users    interface{} `json:"users"`
	Onlines  interface{} `json:"onlines"`
}

// OnlineUser is the data structure of online user
type OnlineUser struct {
	UID int    `json:"userId"`
	IP  string `json:"ip"`
}

// UserTraffic is the data structure of traffic
type UserTraffic struct {
	UID      int    `json:"id"`
	Upload   int64  `json:"up"`
	Download int64  `json:"down"`
	Ip       string `json:"ip"`
}

// NodeStatus Node status report
type NodeStatus struct {
	CPU    float64 `json:"cpu"`
	Mem    float64 `json:"mem"`
	Net    string  `json:"net"`
	Disk   float64 `json:"disk"`
	Uptime uint64  `json:"uptime"`
}

type RuleItem struct {
	ID      int    `json:"id"`
	Content string `json:"regex"`
}

type IllegalItem struct {
	ID  int `json:"list_id"`
	UID int `json:"user_id"`
}
