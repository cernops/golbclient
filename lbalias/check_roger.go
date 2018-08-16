package lbalias

import (
	"os"
	"regexp"

	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const ROGER_CURRENT_FILE = "/etc/roger/current.yaml"

func getRogerState(lbalias *LBalias) string {
	myhost, err := os.Hostname()

	if err != nil {
		panic(err)
	}
	fullName, _ := regexp.Compile(".cern.ch$")
	if !fullName.MatchString(myhost) {
		myhost += ".cern.ch"
	}

	//myhost="esperfmons01.cern.ch"
	url := "http://woger-direct.cern.ch:9098/roger/v1/state/" + myhost
	lbalias.DebugMessage("Ready to call roger")
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	state := ""
	if (err == nil) && (resp.StatusCode == 200) {

		var data map[string]interface{}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("There is an error now")
		}

		err = json.Unmarshal([]byte(body), &data)
		if err != nil {
			panic(err)
		}
		state, _ = data["appstate"].(string)
		return state
	}
	lbalias.DebugMessage("[check_roger] caught exception. Trying cached roger appstate", err)
	return get_roger_fact(lbalias)
}

func checkRoger(lbalias *LBalias, line string) bool {
	if lbalias.RogerState == "" {
		lbalias.RogerState = getRogerState(lbalias)
	}

	if lbalias.RogerState == "production" {
		lbalias.DebugMessage("[check_roger] roger appstate is 'production'")

		return false
	}
	if lbalias.RogerState == "ignore_roger" {
		return false
	}

	lbalias.DebugMessage("[check_roger] Node will go out of LB alias because roger appstate is '" + lbalias.RogerState + "' that is different from 'production'")
	return true

}

func get_roger_fact(lbalias *LBalias) string {

	f, err := os.Open(ROGER_CURRENT_FILE)
	if err != nil {
		fmt.Println("Can't read file "+ROGER_CURRENT_FILE, err)
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lbalias.DebugMessage("[roger_fact] Checking the roger facts")

	state, _ := regexp.Compile("^appstate: *([^ \t\n]+)")
	for scanner.Scan() {
		line := scanner.Text()
		match := state.FindStringSubmatch(line)
		if len(match) > 0 {
			lbalias.DebugMessage("[check_roger] cached appstate is " + match[1])
			return match[1]
		}
	}
	lbalias.DebugMessage("[check_roger] cached appstate is None. Ignoring roger appstate")
	return "ignore_roger"
}
