package huomao

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	getFlashReg = regexp.MustCompile(`getFlash\("([0-9]+)","([0-9a-zA-z]+)","([0-9]+)"\);`)
	stateReg    = regexp.MustCompile(`is_live = "([0-9]+)";`)
)

type ChatData struct {
	Host  string `json:"host"`
	Port  string `json:"port"`
	Token string `json:"token"`
	Uid   string `json:"uid"`
	Group int    `json:"group"`
}

type ChatInfo struct {
	Code    string   `json:"code"`
	Status  bool     `json:"status"`
	Message string   `json:"message"`
	Data    ChatData `json:"data"`
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

func GetBarrageLiveStateRoomId(uri string) (string, string, error) {

	body, err := doHttp(uri)
	if err != nil {
		return "", "", err
	}
	fmt.Println(string(body))
	bs := string(body)
	matchs := getFlashReg.FindStringSubmatch(bs)
	roomId := matchs[1]
	fmt.Println(matchs)
	matchs = stateReg.FindStringSubmatch(bs)
	return matchs[1], roomId, nil
}

func GetBarrageChatInfo(roomStr string) (*ChatInfo, error) {
	km := url.Values{}
	km.Add("callback", "jQuery171032695039477104815_1477741089191")
	km.Add("cid", roomStr)
	km.Add("_", strconv.Itoa(int(time.Now().Unix()*100)))

	uri := "http://chat.huomao.com/chat/getToken?" + km.Encode()
	req, err := http.NewRequest("GET", uri, nil)
	client := &http.Client{}
	req.Header.Add("User-Agent", "User-Agent: Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(string(body))
	chatInfoStr := strings.TrimPrefix(string(body), "jQuery171032695039477104815_1477741089191(")
	chatInfoStr = strings.TrimSuffix(chatInfoStr, ")")
	var chatInfo ChatInfo
	err = json.Unmarshal([]byte(chatInfoStr), &chatInfo)
	if err != nil {
		return nil, err
	}

	return &chatInfo, nil
}
