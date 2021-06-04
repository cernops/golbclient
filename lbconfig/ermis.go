package lbconfig

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	logger "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ClientAuth struct {
	URL         string
	CACert      string
	HostCert    string
	HostKey     string
	AuthTimeout int
	Status      []*status
	Client      *http.Client
}
type status struct {
	AliasName string
	Secret    string
	Load      int
}

// getConn prefills an instance of the ClientAuth struct with the config values
func getConn(path string) (*ClientAuth, error) {
	// Create config structure
	conn := &ClientAuth{}
	//Validate its a readable file
	if err := validateConfigFile(path); err != nil {
		return nil, err
	}
	// Open config file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)
	// Start YAML decoding from file
	if err := d.Decode(&conn); err != nil {
		return nil, err
	}

	return conn, nil
}

// ValidateConfigFile just makes sure, that the path provided is a file,
// that can be read
func validateConfigFile(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		e := fmt.Errorf("provided filepath for the config file is a directory")
		return e
	}
	return nil
}

//InitConnection initiates a new connection with teigi
func (c *ClientAuth) initConnection() error {
	caCert, err := ioutil.ReadFile(c.CACert)
	if err != nil {
		logger.Error(err)

	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(c.HostCert, c.HostKey)
	if err != nil {
		logger.Error(err)

	}

	c.Client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:       caCertPool,
				Certificates:  []tls.Certificate{cert},
				Renegotiation: tls.RenegotiateOnceAsClient,
			},
		},
		Timeout: time.Duration(time.Duration(c.AuthTimeout) * time.Second),
	}
	return nil
}

func (l *AppLauncher) PostToErmis(configpath string) (int, error) {
	type reply struct {
		Message string
	}
	var (
		msg reply
	)

	conn, err := getConn(configpath)
	if err != nil {
		logger.Error(err)
	}
	if err := conn.initConnection(); err != nil {
		logger.Errorf("Error while initiating the connection %v , error:", conn.URL, err.Error())
	}
    //split the calculated alias1:load1,,,aliasN:loadN string
	aliasesSplitted := strings.Split(l.PostErmis, ",")
    
	//verify lbpost.yaml and lbaliases file contents
	if conn.verifyConfigs(aliasesSplitted) != 0{
		 logger.Fatal("misconfigured config files")

	}
	//set load value in the Status struct(alias name and secret are already set from lbpost.yaml)
	if err := conn.setload(aliasesSplitted); err != nil {
		logger.Errorf("failed to update load with the latest value, before dispatching POST request, error:%v", err) 
	}
	injson, err := json.Marshal(conn.Status)
	if err != nil {
		logger.Fatalf("could not marshal struct %v into json with error: %v", conn.Status, err)
	}

	request, err := http.NewRequest("POST", conn.URL,
		bytes.NewBuffer(injson))

	if err != nil {
		logger.Fatalf("failed to prepare POST request with error: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	resp, err := conn.Client.Do(request)
	if err != nil {
		logger.Fatalf("failed to dispatch POST request to %v with error: %v", conn.URL, err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("error reading the Body of response from goermis")

	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		logger.Errorf("user not authorized, received status code %v ", resp.StatusCode)

	}

	if err = json.Unmarshal(data, &msg); err != nil {
		logger.Errorf("error on unmarshalling response from goermis, error %v ", err)
	}
	logger.Debugf(msg.Message)
	//return values are useful while testing
	return resp.StatusCode, err
}

//setload sets the value of ClientAuth.Status.Load with the new calculated one
func (c *ClientAuth) setload(aliases []string) error{
		for _, alias := range aliases {
			name := strings.Split(alias, "=")[0]
			load, err := strconv.Atoi(strings.Split(alias, "=")[1])
			if err !=nil{
				return err 
			}
			for _, status := range c.Status {
				if name == status.AliasName {
					status.Load = load
				}
			}
		}
		return nil
}
//verifyConfigs checks the correct configuration of aliases and their secrets in lbaliases and lbpost.yaml files
func (c *ClientAuth)verifyConfigs(lbaliases []string) int {
	var( 
		diff               []string
		aliasesInlbpost    []string
		aliasesInlbaliases []string
		
	)
	
    //if the number of aliases is different, return false immediately
	if len(lbaliases) != len(c.Status) {
		logger.Debug("misconfiguration for alias definition in the files lbaliases and lbpost.yaml, missing alias(es)")
		return 1
	}

    //get rid of the load value
	for _,v:= range lbaliases{
		aliasesInlbaliases = append(aliasesInlbaliases, strings.Split(v,"=")[0])
	}

	for _, v := range c.Status{
		 //make sure all aliases have secrets defined
		 if v.Secret == ""{
			logger.Debug("missing secret for alias %v in lbpost.yaml file",v.AliasName)
			return 1
		}
		//extract the declared names in lbpost.yaml
		aliasesInlbpost = append(aliasesInlbpost, v.AliasName)


	}
	
	
	//check the aliases in lbaliases file are the same as the one in lbpost.yaml
    for i:=0; i<2;i++{
		for _, v1 := range aliasesInlbaliases {
			found := false
			for _, v2 := range aliasesInlbpost {
				if v1 == v2{
					found = true
					break
					}
				}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, v1)
				}
			}
			//change their place , so that on the second itteration we check the second file for missing aliases
			if i ==0{
				aliasesInlbaliases, aliasesInlbpost = aliasesInlbpost, aliasesInlbaliases
			}

	}
	logger.Tracef("Content difference between lbpost.yaml and lbaliases %v", diff)
    return len(diff)
}
	
