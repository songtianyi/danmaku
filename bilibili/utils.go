package bilibili

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
)

var (
	roomReg   = regexp.MustCompile("var ROOMID = (\\d+)")
	serverReg = regexp.MustCompile("<server>(.*?)</server>")
	stateReg  = regexp.MustCompile("<state>(.*?)</state>")
)

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
	body, err := doHttp(uri)
	if err != nil {
		return "", err
	}
	matchs := roomReg.FindStringSubmatch(string(body))
	if len(matchs) < 2 {
		return "", fmt.Errorf("ROOMID submatch %q", matchs)
	}
	return matchs[1], nil
}

func GetBarrageServerAndLiveState(room string) (string, string, error) {
	urii := "http://live.bilibili.com/api/player?id=cid:" + room
	body, err := doHttp(urii)
	if err != nil {
		return "", "", err
	}
	matchs := serverReg.FindStringSubmatch(string(body))
	if len(matchs) < 2 {
		return "", "", fmt.Errorf("server submatch %q", matchs)
	}
	server := matchs[1]

	matchs = stateReg.FindStringSubmatch(string(body))
	if len(matchs) < 2 {
		return "", "", fmt.Errorf("state submatch %q", matchs)
	}
	state := matchs[1]
	return server, state, nil
}

func RandUser() int {
	return rand.Intn(4e7-1e5) + 1e5
}
