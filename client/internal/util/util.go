package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"
)

const timeLayout = "Jan 2 2006 15.04.05 -0700 MST"

var encodingUnencodedTokens = []string{"%", ":", "[", "]", ",", "\""}
var encodingEncodedTokens = []string{"%25", "%3A", "%5B", "%5D", "%2C", "%22"}
var decodingUnencodedTokens = []string{":", "[", "]", ",", "\"", "%"}
var decodingEncodedTokens = []string{"%3A", "%5B", "%5D", "%2C", "%22", "%25"}

// Client dupa
type Client struct {
	Connection net.Conn
	Username   string
	Room       string
	ignoring   []string
	Properties Properties
}

// Close the client connection and clenup
func (client *Client) Close(doSendMessage bool) {
	if doSendMessage {
		SendClientMessage("disconnect", "", client, false, client.Properties)
	}
	client.Connection.Close()
	clients = removeEntry(client, clients)
}

// Register the connection and cache it
func (client *Client) Register() {
	clients = append(clients, client)
}

// Ignore dupa
func (client *Client) Ignore(username string) {
	client.ignoring = append(client.ignoring, username)
}

// IsIgnoring dupa
func (client *Client) IsIgnoring(username string) bool {
	for _, value := range client.ignoring {
		if value == username {
			return true
		}
	}
	return false
}

// Action dupa
type Action struct {
	Command   string `json:"command"`
	Content   string `json:"content"`
	Username  string `json:"username"`
	IP        string `json:"ip"`
	Timestamp string `json:"timestamp"`
}

// Properties dupa
type Properties struct {
	Hostname                  string
	Port                      string
	JSONEndpointPort          string
	HasEnteredTheRoomMessage  string
	HasLeftTheRoomMessage     string
	HasEnteredTheLobbyMessage string
	HasLeftTheLobbyMessage    string
	ReceivedAMessage          string
	IgnoringMessage           string
	LogFile                   string
}

var actions = []Action{}
var config = Properties{}
var clients []*Client

// LoadConfig dupa
func LoadConfig() Properties {
	if config.Port != "" {
		return config
	}
	pwd, _ := os.Getwd()

	payload, err := ioutil.ReadFile(pwd + "/config.json")
	CheckForError(err, "Unable to read config file")

	var dat map[string]interface{}
	err = json.Unmarshal(payload, &dat)
	CheckForError(err, "Invalid JSON in config file")

	var rtn = Properties{
		Hostname:                  dat["Hostname"].(string),
		Port:                      dat["Port"].(string),
		JSONEndpointPort:          dat["JSONEndpointPort"].(string),
		HasEnteredTheRoomMessage:  dat["HasEnteredTheRoomMessage"].(string),
		HasLeftTheRoomMessage:     dat["HasLeftTheRoomMessage"].(string),
		HasEnteredTheLobbyMessage: dat["HasEnteredTheLobbyMessage"].(string),
		HasLeftTheLobbyMessage:    dat["HasLeftTheLobbyMessage"].(string),
		ReceivedAMessage:          dat["ReceivedAMessage"].(string),
		IgnoringMessage:           dat["IgnoringMessage"].(string),
		LogFile:                   dat["LogFile"].(string),
	}
	config = rtn
	return rtn
}

func removeEntry(client *Client, arr []*Client) []*Client {
	rtn := arr
	index := -1
	for i, value := range arr {
		if value == client {
			index = i
			break
		}
	}

	if index >= 0 {
		rtn = make([]*Client, len(arr)-1)
		copy(rtn, arr[:index])
		copy(rtn[index:], arr[index+1:])
	}

	return rtn
}

// SendClientMessage dupa
func SendClientMessage(messageType string, message string, client *Client, thisClientOnly bool, props Properties) {

	if thisClientOnly {
		message = fmt.Sprintf("/%v", messageType)
		fmt.Fprintln(client.Connection, message)

	} else if client.Username != "" {
		LogAction(messageType, message, client, props)

		payload := fmt.Sprintf("/%v [%v] %v", messageType, client.Username, message)

		for _, _client := range clients {
			if (thisClientOnly && _client.Username == client.Username) ||
				(!thisClientOnly && _client.Username != "") {

				if messageType == "message" && client.Room != _client.Room || _client.IsIgnoring(client.Username) {
					continue
				}

				fmt.Fprintln(_client.Connection, payload)
			}
		}
	}
}

// CheckForError dupa
func CheckForError(err error, message string) {
	if err != nil {
		println(message+": ", err.Error())
		os.Exit(1)
	}
}

// EncodeCSV dupa
func EncodeCSV(value string) string {
	return strings.Replace(value, "\"", "\"\"", -1)
}

// Encode dupa
func Encode(value string) string {
	return replace(encodingUnencodedTokens, encodingEncodedTokens, value)
}

// Decode dupa
func Decode(value string) string {
	return replace(decodingEncodedTokens, decodingUnencodedTokens, value)
}

// replace dupa
func replace(fromTokens []string, toTokens []string, value string) string {
	for i := 0; i < len(fromTokens); i++ {
		value = strings.Replace(value, fromTokens[i], toTokens[i], -1)
	}
	return value
}

// LogAction dupa
func LogAction(action string, message string, client *Client, props Properties) {
	ip := client.Connection.RemoteAddr().String()
	timestamp := time.Now().Format(timeLayout)

	actions = append(actions, Action{
		Command:   action,
		Content:   message,
		Username:  client.Username,
		IP:        ip,
		Timestamp: timestamp,
	})

	if props.LogFile != "" {
		if message == "" {
			message = "N/A"
		}
		fmt.Printf("logging values %s, %s, %s\n", action, message, client.Username)

		logMessage := fmt.Sprintf("\"%s\", \"%s\", \"%s\", \"%s\", \"%s\"\n",
			EncodeCSV(client.Username), EncodeCSV(action), EncodeCSV(message),
			EncodeCSV(timestamp), EncodeCSV(ip))

		f, err := os.OpenFile(props.LogFile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			err = ioutil.WriteFile(props.LogFile, []byte{}, 0600)
			f, err = os.OpenFile(props.LogFile, os.O_APPEND|os.O_WRONLY, 0600)
			CheckForError(err, "Cant create log file")
		}

		defer f.Close()
		_, err = f.WriteString(logMessage)
		CheckForError(err, "Can't write to log file")
	}
}

// QueryMessages dupa
func QueryMessages(actionType string, search string, username string) []Action {

	isMatch := func(action Action) bool {
		if actionType != "" && action.Command != actionType {
			return false
		}
		if search != "" && !strings.Contains(action.Content, search) {
			return false
		}
		if username != "" && action.Username != username {
			return false
		}
		return true
	}

	rtn := make([]Action, 0, len(actions))

	for _, value := range actions {
		if isMatch(value) {
			rtn = append(rtn, value)
		}
	}

	return rtn
}
