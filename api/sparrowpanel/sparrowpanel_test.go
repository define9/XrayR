package sparrowpanel_test

import (
	"fmt"
	"testing"

	"github.com/XrayR-project/XrayR/api"
	"github.com/XrayR-project/XrayR/api/sparrowpanel"
)

func CreateV2rayClient() api.API {
	apiConfig := &api.Config{
		APIHost:  "http://localhost:60518/",
		Key:      "123456",
		NodeID:   1,
		NodeType: "V2ray",
	}
	client := sparrowpanel.New(apiConfig)
	return client
}

func CreateClient(nodeType string) api.API {
	apiConfig := &api.Config{
		APIHost:  "http://localhost:60518/",
		Key:      "123456",
		NodeID:   1,
		NodeType: nodeType,
	}
	client := sparrowpanel.New(apiConfig)
	return client
}

func TestGetV2rayNodeinfo(t *testing.T) {
	client := CreateClient("V2ray")
	client.Debug()
	nodeInfo, err := client.GetNodeInfo()
	if err != nil {
		t.Error(err)
	}
	t.Log(nodeInfo)
}

func TestGetSSNodeinfo(t *testing.T) {
	client := CreateClient("Shadowsocks")
	client.Debug()
	nodeInfo, err := client.GetNodeInfo()
	if err != nil {
		t.Error(err)
	}
	t.Log(nodeInfo)
}

func TestGetTrojanNodeinfo(t *testing.T) {
	client := CreateClient("Trojan")
	client.Debug()
	nodeInfo, err := client.GetNodeInfo()
	if err != nil {
		t.Error(err)
	}
	t.Log(nodeInfo)
}

func TestGetSSinfo(t *testing.T) {
	client := CreateClient("")

	nodeInfo, err := client.GetNodeInfo()
	if err != nil {
		t.Error(err)
	}
	t.Log(nodeInfo)
}

func TestGetUserList(t *testing.T) {
	client := CreateV2rayClient()

	userList, err := client.GetUserList()
	if err != nil {
		t.Error(err)
	}

	t.Log(userList)
}

func TestReportNodeStatus(t *testing.T) {
	client := CreateV2rayClient()
	nodeStatus := &api.NodeStatus{
		CPU: 1, Mem: 1, Disk: 1, Uptime: 256,
	}
	err := client.ReportNodeStatus(nodeStatus)
	if err != nil {
		t.Error(err)
	}
}

func TestReportReportNodeOnlineUsers(t *testing.T) {
	client := CreateV2rayClient()
	userList, err := client.GetUserList()
	if err != nil {
		t.Error(err)
	}

	onlineUserList := make([]api.OnlineUser, len(*userList))
	for i, userInfo := range *userList {
		onlineUserList[i] = api.OnlineUser{
			UID: userInfo.UID,
			IP:  fmt.Sprintf("1.1.1.%d", i),
		}
	}
	// client.Debug()
	err = client.ReportNodeOnlineUsers(&onlineUserList)
	if err != nil {
		t.Error(err)
	}
}

func TestReportReportUserTraffic(t *testing.T) {
	client := CreateV2rayClient()
	userList, err := client.GetUserList()
	if err != nil {
		t.Error(err)
	}
	generalUserTraffic := make([]api.UserTraffic, len(*userList))
	for i, userInfo := range *userList {
		generalUserTraffic[i] = api.UserTraffic{
			UID:      userInfo.UID,
			Upload:   114514,
			Download: 114514,
		}
	}
	// client.Debug()
	err = client.ReportUserTraffic(&generalUserTraffic)
	if err != nil {
		t.Error(err)
	}
}

func TestGetNodeRule(t *testing.T) {
	client := CreateV2rayClient()

	ruleList, err := client.GetNodeRule()
	if err != nil {
		t.Error(err)
	}

	t.Log(ruleList)
}

func TestReportIllegal(t *testing.T) {
	client := CreateV2rayClient()

	detectResult := []api.DetectResult{
		{1, 2},
		{1, 3},
	}
	client.Debug()
	err := client.ReportIllegal(&detectResult)
	if err != nil {
		t.Error(err)
	}
}
