package panda

import (
	"encoding/json"
	"fmt"
	"github.com/songtianyi/rrframework/config"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ChatData struct {
	AppId        string   `json:"appid"`
	Rid          int      `json:"rid"`
	Sign         string   `json:"sign"`
	AuthType     string   `json:"authType"`
	Ts           int      `json:"ts"`
	ChatAddrList []string `json:"chat_addr_list"`
}

type ChatInfo struct {
	Errno  int      `json:"errno"`
	Errmsg string   `json:"errmsg"`
	Data   ChatData `json:"data"`
}

func doHttp(uri string) ([]byte, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return body, nil
}

func GetRoomId(uri string) (string, error) {
	uri = strings.Trim(uri, "/")
	it := strings.Split(uri, "/")
	if len(it) < 2 {
		return "", fmt.Errorf("url %s not valid", uri)
	}
	roomStr := it[len(it)-1]
	return roomStr, nil
}

func GetBarrageLiveState(roomStr string) (string, error) {

	km := url.Values{}
	km.Add("roomid", roomStr)
	km.Add("pub_key", "")
	km.Add("_", strconv.Itoa(int(time.Now().Unix())))

	api := "http://www.panda.tv/api_room?" + km.Encode()
	body, err := doHttp(api)
	if err != nil {
		return "", err
	}
	jc, err := rrconfig.LoadJsonConfigFromBytes(body)
	if err != nil {
		return "", err
	}
	status, _ := jc.GetString("data.videoinfo.status")
	return status, nil
}

func GetBarrageChatInfo(roomStr string) (*ChatInfo, error) {
	km := url.Values{}
	km.Add("roomid", roomStr)
	km.Add("_", strconv.Itoa(int(time.Now().Unix())))

	api := "http://www.panda.tv/ajax_chatinfo?" + km.Encode()
	body, err := doHttp(api)
	if err != nil {
		return nil, err
	}
	var chatInfo ChatInfo
	err = json.Unmarshal(body, &chatInfo)
	if err != nil {
		return nil, err
	}

	km = url.Values{}
	km.Add("rid", strconv.Itoa(chatInfo.Data.Rid))
	km.Add("roomid", roomStr)
	km.Add("retry", "0")
	km.Add("sign", chatInfo.Data.Sign)
	km.Add("ts", strconv.Itoa(chatInfo.Data.Ts))
	km.Add("_", strconv.Itoa(int(time.Now().Unix())))

	api = "http://api.homer.panda.tv/chatroom/getinfo?" + km.Encode()
	body, err = doHttp(api)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &chatInfo)
	if err != nil {
		return nil, err
	}
	return &chatInfo, nil
}
