package main

import (
	"bytes"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/plugin"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type BasicPlugin struct{}

type ServiceKeyUsers struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type ServiceKeyUrls struct {
	Gfsh string `json:"gfsh"`
}

type ServiceKey struct {
	Urls  ServiceKeyUrls    `json:"urls"`
	Users []ServiceKeyUsers `json:"users"`
}

type ClusterManagementResult struct {
	StatusCode string `json:"statusCode"`
	StatusMessage string `json:"statusMessage"`
	MemberStatus []MemberStatus `json:"memberStatus"`
	Result []map[string]interface{} `json:"result"`
}

type MemberStatus struct {
	ServerName string
	Success bool
	Message string
}

const missingInformationMessage string = `Your request was denied.
You are missing a username, password, or the correct endpoint.
`
const incorrectUserInputMessage string = `Your request was denied.
The format of your request is incorrect.

For help see: cf gf --help`
const invalidPCCInstanceMessage string = `You entered %s which not a deployed PCC instance.
To deploy this as an instance, enter: 

	cf create-service p-cloudcache <region_plan> %s

For help see: cf create-service --help

`
const noServiceKeyMessage string = `Please create a service key for %s.
To create a key enter: 

	cf create-service-key %s <your_key_name>
	
For help see: cf create-service-key --help

`

func getServiceKeyFromPCCInstance(pccService string) (serviceKey string, err error) {
	cmd := exec.Command("cf", "service-keys", pccService)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return "", errors.New(invalidPCCInstanceMessage)
	}
	servKeyOutput := &out
	keysStr := servKeyOutput.String()
	splitKeys := strings.Split(keysStr, "\n")
	hasKey := false
	if strings.Contains(splitKeys[1], "No service key for service instance"){
		return "", errors.New(noServiceKeyMessage)
	}
	for _, value := range splitKeys {
		line := strings.Fields(value)
		if len(line) > 0 {
			if hasKey {
				serviceKey = line[0]
				return
			} else if line[0] == "name" {
				hasKey = true
			}
		}
	}
	if serviceKey == "" {
		return serviceKey, errors.New(noServiceKeyMessage)
	}
	return
}

func getUsernamePasswordEndpoint(pccService string, key string) (username string, password string, endpoint string) {
	username = ""
	password = ""
	endpoint = ""
	cmd := exec.Command("cf", "service-key", pccService, key)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	servKeyOutput := &out
	keyInfo := servKeyOutput.String()
	splitKeyInfo := strings.Split(keyInfo, "\n")
	splitKeyInfo = splitKeyInfo[2:] //take out first two lines of cf service-key ... output
	joinKeyInfo := strings.Join(splitKeyInfo, "\n")

	serviceKey := ServiceKey{}

	err = json.Unmarshal([]byte(joinKeyInfo), &serviceKey)
	if err != nil {
		log.Fatal(err)
	}
	endpoint = serviceKey.Urls.Gfsh
	endpoint = strings.Replace(endpoint, "gemfire/v1", "geode-management/v2", 1)
	for _ , user := range serviceKey.Users {
		if strings.HasPrefix(user.Username, "cluster_operator") {
			username = user.Username
			password = user.Password
		}
	}
	return
}

func ValidatePCCInstance(ourPCCInstance string, pccInstancesAvailable []string) (error){
	for _, pccInst := range pccInstancesAvailable {
		if ourPCCInstance == pccInst {
			return nil
		}
	}
	return errors.New(invalidPCCInstanceMessage)
}

func getCompleteEndpoint(endpoint string, clusterCommand string) (string){
	urlEnding := ""
	if clusterCommand == "list-regions"{
		urlEnding = "/regions"
	} else if clusterCommand == "list-members"{
		urlEnding = "/members"
	}
	endpoint = endpoint + urlEnding
	return endpoint
}

func getUrlOutput(endpointUrl string, username string, password string) (urlResponse string){
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", endpointUrl, nil)
	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil{
		log.Fatal(err)
	}
	respInAscii, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil{
		log.Fatal(1)
	}
	urlResponse = fmt.Sprintf("%s", respInAscii)
	return
}

func Fill(columnSize int, value string, filler string) (response string){
	if len(value) > columnSize - 1{
		response = " " + value[:columnSize-1]
		return
	}
	numFillerChars := columnSize - len(value) - 1
	response = " " + value + strings.Repeat(filler, numFillerChars)
	return
}


func getTableHeadersFromClusterCommand(clusterCommand string) (tableHeaders []string){
	if clusterCommand == "list-regions"{
		tableHeaders = append(tableHeaders, "name", "type", "groups", "entryCount", "regionAttributes")
	} else if clusterCommand =="list-members"{
		tableHeaders = append(tableHeaders, "id", "host", "status", "pid")
	} else{
		return
	}
	return
}

func GetAnswerFromUrlResponse(clusterCommand string, urlResponse string) (response string){
	urlOutput := ClusterManagementResult{}
	err := json.Unmarshal([]byte(urlResponse), &urlOutput)
	if err != nil {
		log.Fatal(err)
	}
	response = "Status Code: " + urlOutput.StatusCode + "\n"
	if urlOutput.StatusMessage != ""{
		response += "Status Message: " + urlOutput.StatusMessage + "\n"
	}
	response += "\n"

	tableHeaders := getTableHeadersFromClusterCommand(clusterCommand)
	for _, header := range tableHeaders {
		response += Fill(20, header, " ") + "|"
	}
	response += "\n" + Fill (20 * len(tableHeaders) + 5, "", "-") + "\n"

	memberCount := 0
	for _, result := range urlOutput.Result{
		memberCount++
		for _, key := range tableHeaders {
			if result[key] == nil {
				response += Fill(20, "", " ") + "|"
			} else {
				resultVal := result[key]
				if fmt.Sprintf("%T", result[key]) == "float64"{
					resultVal = fmt.Sprintf("%.0f", result[key])
				}
				response += Fill(20, fmt.Sprintf("%s",resultVal), " ") + "|"
			}
		}
		response += "\n"
	}
	if clusterCommand == "list-regions"{
		response += "\nNumber of Regions: " + strconv.Itoa(memberCount)
	} else if clusterCommand == "list-members"{
		response += "\nNumber of Members: " + strconv.Itoa(memberCount)
	}
	return
}


func GetJsonFromUrlResponse(urlResponse string) (jsonOutput string){
	urlOutput := ClusterManagementResult{}
	err := json.Unmarshal([]byte(urlResponse), &urlOutput)
	if err != nil {
		log.Fatal(err)
	}
	jsonExtracted, err := json.MarshalIndent(urlOutput, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	jsonOutput = string(jsonExtracted)
	return
}

func isRegionInGroups(regionInGroup bool, groupsWeHave interface{}, groupsWeWant []string) (isInGroups bool){
	if regionInGroup{
		isInGroups = true
		return
	}
	for _, group := range groupsWeWant{
		resultValToStr := fmt.Sprintf("%s", groupsWeHave)
		for  _, regionName := range strings.Split(resultValToStr[1:len(resultValToStr)-1], " "){
			if regionName == group{
				isInGroups = true
				return
			}
		}
	}
	isInGroups = false
	return
}

func EditResponseOnGroup(urlResponse string, groups []string, clusterCommand string) (editedUrlResponse string){
	urlOutput := ClusterManagementResult{}
	err := json.Unmarshal([]byte(urlResponse), &urlOutput)
	if err != nil {
		log.Fatal(err)
	}
	var newUrlOutputResult string
	var newResult []map[string]interface{}
	for _, result := range urlOutput.Result{
		regionInGroups:= false
		for _, key :=range getTableHeadersFromClusterCommand(clusterCommand){
			if key == "groups"{
				regionInGroups = isRegionInGroups(regionInGroups, result[key], groups)
			}
			if regionInGroups{
				break
			}
		}
		if regionInGroups{
			newUrlOutputResult += fmt.Sprintf("%s",result)
			newResult = append(newResult, result)
		}
	}
	urlOutput.Result = newResult
	byteResponse, err := json.Marshal(urlOutput)
	if err != nil {
		log.Fatal(err)
	}
	editedUrlResponse = string(byteResponse)
	return
}

func (c *BasicPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	start := time.Now()
	if args[0] == "CLI-MESSAGE-UNINSTALL"{
		return
	}
	var username, password, endpoint, pccInUse, clusterCommand, serviceKey string
	var groups []string
	if len(args) >= 3 {
		pccInUse = args[2]
		clusterCommand = args[1]
	} else{
		fmt.Println(incorrectUserInputMessage)
		return
	}
	if os.Getenv("CFLOGIN") != "" && os.Getenv("CFPASSWORD") != "" && os.Getenv("CFENDPOINT") != "" {
		username = os.Getenv("CFLOGIN")
		password = os.Getenv("CFPASSWORD")
		endpoint = os.Getenv("CFENDPOINT")
	} else {
		var err error
		serviceKey, err = getServiceKeyFromPCCInstance(pccInUse)
		if err != nil{
			fmt.Printf(err.Error(), pccInUse, pccInUse)
			os.Exit(1)
		}
		username, password, endpoint = getUsernamePasswordEndpoint(pccInUse, serviceKey)
	}

	endpoint = getCompleteEndpoint(endpoint, clusterCommand)
	urlResponse := getUrlOutput(endpoint, username, password)
	hasJ := false
	for _, arg := range args {
		if strings.HasPrefix(arg, "-g="){
			groups = strings.Split(arg[3:], ",")
			urlResponse = EditResponseOnGroup(urlResponse, groups, clusterCommand)
		}
		if arg == "-j"{
			hasJ = true
		}
	}
	if hasJ{
		fmt.Println(GetJsonFromUrlResponse(urlResponse))
		return
	}
	fmt.Println("PCC in use: " + pccInUse)
	fmt.Println("Service key: " + serviceKey)

	if username != "" && password != "" && clusterCommand != "" && endpoint != "" {
		answer := GetAnswerFromUrlResponse(clusterCommand, urlResponse)
		fmt.Println()
		fmt.Println(answer)
		fmt.Println()
	} else {
		fmt.Println(missingInformationMessage)
	}
	t := time.Now()
	fmt.Println(t.Sub(start))
}


func (c *BasicPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "GF_InDev",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 0,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 7,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "gf",
				HelpText: "gf's help text",
				UsageDetails: plugin.Usage{
					Usage: "   cf gf [action] [pcc_instance] [*options] (* = optional)\n" +
						"	Actions: \n" +
						"		list-regions, list-members\n" +
						"	Options: \n" +
						"		-h : this help screen\n" +
						"		-j : json output of API endpoint\n" +
						"		-g : followed by group(s), split by comma, only data within those groups\n" +
						"			(example: cf gf list-regions --g=group1,group2)",
				},
			},
		},
	}
}


func main() {
	plugin.Start(new(BasicPlugin))
}