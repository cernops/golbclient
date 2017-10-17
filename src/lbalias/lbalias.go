package lbalias

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	//"log"

	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
)

const ACCEPTABLE_BLOCK_RATE = 0.90
const ACCEPTABLE_INODE_RATE = 0.95
const ROGER_CURRENT_FILE = "/etc/roger/current.yaml"
const LEMON_CLI = "/usr/sbin/lemon-cli"

type LBalias struct {
	Name           string
	ConfigFile     string
	Debug          bool
	NoLogin        bool
	Syslog         bool
	GData          []string
	MData          []string
	CheckXsessions int
	RogerState     string
	//	Metric     string
	//	CheckSwapping bool
	//	CheckXsessions bool
	CheckMetricList []MetricEntry
	//	LoadMetricList []int
	//	ConstantList []int
}
type LBcheck struct {
	Code     int
	Function func(*LBalias, string) bool
}

var allLBchecks = map[string]LBcheck{
	"NOLOGIN":       LBcheck{1, checkNoLogin},
	"TMPFULL":       LBcheck{6, checkTmpFull},
	"SSHDAEMON":     LBcheck{7, checkDaemon(22)},
	"WEBDAEMON":     LBcheck{8, checkDaemon(80)},
	"FTPDAEMON":     LBcheck{9, checkDaemon(21)},
	"AFS":           LBcheck{10, checkAFS},
	"GRIDFTPDAEMON": LBcheck{11, checkDaemon(2811)},
	"LEMON":         LBcheck{12, checkLemon},
	"ROGER":         LBcheck{13, checkRoger},
	//The rest are a bit special. They don't return immediately
	"XSESSIONS": LBcheck{0, checkXsession}}

type MetricEntry struct {
	PrefixOp, Metric, Index, Op, Value string
}

//These are all the current tests
//
//

// The lemon metrics are done in two steps: The first one is to add all of them to the configuration
// The second step is to call all of them in one go
func checkLemon(lbalias *LBalias, line string) bool {

	lbalias.DebugMessage("[add_lemon_check] Adding Lemon metric ", line)

	actions, _ := regexp.Compile("(?i)^CHECK LEMON ([-])?(_)?([0-9]+)(:[0-9]+)?([^0-9]+)([0-9]*.?[0-9]*)")

	found := actions.FindStringSubmatch(line)
	if len(found) > 0 {
		prefix_op, underscore, metric, index, op, value := found[1], found[2], found[3], found[4], found[5], found[6]
		if underscore != "_" {
			fmt.Printf("[add_lemon_check] Invalid metric.  Must start with _ (", found[0], ")")
			return true
		}
		if index == "" {
			index = "1"
		}
		lbalias.DebugMessage("[add_lemon_check] prefix=", prefix_op, ", metric=", metric, ", index=", index, ", op=", op, ", value=", value)
		lbalias.CheckMetricList = append(lbalias.CheckMetricList, MetricEntry{prefix_op, metric, index, op, value})
		//lbalias.MetricList= append(lbalias.MetricList, "")
	} else {
		fmt.Printf("[add_lemon_check] Invalid expresion: ", line)
		return true
	}
	return false
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

func checkXsession(lbalias *LBalias, line string) bool {
	lbalias.DebugMessage("[xsessions] Checking the xsessions")

	lbalias.CheckXsessions = 1
	return false
}
func daemonListening(port int, proc string) bool {
	file, err := os.Open(proc)
	if err != nil {
		fmt.Println("Error openning", proc)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// The format of the file is 'sl  local_address rem_address   st'

	portHex := fmt.Sprintf("%04x", port)

	//fmt.Println("Looking for ", portHex)
	portOpen, _ := regexp.Compile("^ *[0-9]+: [0-9A-F]+:" + portHex + " [0-9A-F]+:[0-9A-F]+ 0A")
	for scanner.Scan() {
		line := scanner.Text()
		if portOpen.MatchString(line) {
			return true
		}
	}
	return false

}
func checkDaemon(port int) func(*LBalias, string) bool {
	return func(lbalias *LBalias, line string) bool {

		if (port < 1) || (port > 65535) {
			return true
		}
		notok := true
		listen := []string{}
		if daemonListening(port, "/proc/net/tcp") {
			notok = false
			listen = append(listen, "ipv4")
		}
		if daemonListening(port, "/proc/net/tcp6") {
			notok = false
			listen = append(listen, "ipv6")
		}

		if lbalias.Debug {
			if len(listen) > 0 {
				for _, p := range listen {
					fmt.Printf("[check_daemon %s] daemon on port %d is listening\n", p, port)

				}
			} else {
				fmt.Printf("[check_daemon] daemon on port %d is not listening\n", port)
			}
		}
		return notok
	}
}

func checkTmpFull(lbalias *LBalias, line string) bool {
	var stat syscall.Statfs_t

	syscall.Statfs("/tmp", &stat)
	blockLevel := 1 - (float64(stat.Bavail) / float64(stat.Blocks))
	inodeLevel := 1 - (float64(stat.Ffree) / float64(stat.Files))
	lbalias.DebugMessage("[check_tmpfull] blocks occupancy: %.2f%% inodes occupancy: %.2f%%\n", blockLevel*100, inodeLevel*100)

	return ((blockLevel > ACCEPTABLE_BLOCK_RATE) || (inodeLevel > ACCEPTABLE_INODE_RATE))
}

func checkAFS(lbalias *LBalias, line string) bool {
	return true
}

func checkNoLogin(lbalias *LBalias, line string) bool {
	if lbalias.NoLogin {
		return false
	}
	nologin := [2]string{"/etc/nologin", "/etc/iss.nologin"}

	if lbalias.Name != "" {
		nologin[1] += "." + lbalias.Name
	}
	for _, file := range nologin {
		_, err := os.Stat(file)

		if err == nil {
			lbalias.DebugMessage("[check_nologin] %s present\n", file)
			return true
		}
	}
	lbalias.DebugMessage("[check_nologin] users allowed to log in\n")
	return false

}

//
// And here we add the methods of the class
//
//

func (lbalias LBalias) DebugMessage(s ...interface{}) {
	if lbalias.Debug {
		fmt.Println(s)
	}

}
func (lbalias LBalias) Evaluate() int {
	lbalias.DebugMessage("[lbalias] Evaluating the alias " + lbalias.Name)

	checks := []string{}
	for key, _ := range allLBchecks {
		checks = append(checks, "("+key+")")
	}
	f, err := os.Open(lbalias.ConfigFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lbalias.DebugMessage("[lbalias] Configuration file opened")

	comment, _ := regexp.Compile("^(#.*)?$")
	actions, _ := regexp.Compile("(?i)^CHECK (" + strings.Join(checks, "|") + ")")
	constant, _ := regexp.Compile("(?i)^LOAD (LEMON)|(CONSTANT) (.*)")
	for scanner.Scan() {
		line := scanner.Text()
		if comment.MatchString(line) {
			continue
		}
		actions := actions.FindStringSubmatch(line)
		if len(actions) > 0 {

			myAction := strings.ToUpper(actions[1])

			if allLBchecks[myAction].Function(&lbalias, line) {
				fmt.Println("THE CHECK OF ", myAction, "FAILED ")
				return -allLBchecks[myAction].Code
			}

			continue
		}
		if constant.MatchString(line) {
			fmt.Println("THERE IS A CONSTANT")
			continue

		}
		fmt.Println("We can't parse the configuration line", line)

	}
	if len(lbalias.CheckMetricList) > 0 {
		if lbalias.checkLemonMetric() {
			lbalias.DebugMessage("[main] Lemon metric check failed")
			return -allLBchecks["LEMON"].Code
		}
	}
	return 0
}

func (lbalias *LBalias) checkLemonMetric() bool {

	var commandArgs = []string{"--script", "-m", ""}
	for _, m := range lbalias.CheckMetricList {
		commandArgs[2] += m.Metric + " "
	}
	lbalias.DebugMessage("Running ", commandArgs)
	out, err := exec.Command(LEMON_CLI, commandArgs...).Output()
	fmt.Println("EXECUTED!", out, err)
	if err != nil {
		fmt.Println("Error executing the lemon cli!", err)
		//return true

	}

	output = "esperfmons01 13163 1508249393 1\nesperfmons01 13423 1508249393 0"

	/*
	   (LcOutput,err) = subprocess.Popen( argList,
	                                      stdout=subprocess.PIPE ).communicate()
	   if err:
	       return 0
	   valuelist = {}
	   lines = LcOutput.split('\n')
	   for line in lines:
	       if len(line) != 0:
	           words = line.split()
	           hostname = words[0]
	           metric = words[1]
	           timestamp = words[2]
	           valuelist[metric] = words[3:]
	   for entry in metriclist:
	       try:
	           valuestring = valuelist[entry.metric][entry.index -1]
	       except KeyError:
	           if debug:
	               printf("Error: Failed to get data for metric=%s\n" % entry.metric)
	           continue
	       if debug:
	           printf("Lemon Metric %s value %s\n",entry.metric,valuestring)
	       value = float(valuestring)
	       op = entry.op
	       metriccheckok=1
	       if entry.prefix_op is not None:
	           if entry.prefix_op == '-':
	               value = -value
	       if debug:
	           printf("Compare %s %f with limit %f\n",op,value,entry.value)
	       if   op == "==":
	           metriccheckok=(value==entry.value)
	       elif (op == "!=") or (op == "<>"):
	           metriccheckok=(value!=entry.value)
	       elif op == ">":
	           metriccheckok=(value>entry.value)
	       elif op == "<":
	           metriccheckok=(value<entry.value)
	       elif op == ">=":
	           metriccheckok=(value>=entry.value)
	       elif op == "<=":
	           metriccheckok=(value<=entry.value)
	       if not metriccheckok:
	           return 1
	   return 0 */
	return false
}
