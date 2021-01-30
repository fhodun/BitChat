package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/fhodun/BitChat/client/internal/util"
)

var standardInputMessageRegex, _ = regexp.Compile(`^\/([^\s]*)\s*(.*)$`)
var chatServerResponseRegex, _ = regexp.Compile(`^\/([^\s]*)\s?(?:\[([^\]]*)\])?\s*(.*)$`)

type Command struct {
	Command, Username, Body string
}

func main() {
	username, properties := getConfig()

	conn, err := net.Dial("tcp", properties.Hostname+":"+properties.Port)
	util.CheckForError(err, "Connection refused")
	defer conn.Close()

	go watchForConnectionInput(username, properties, conn)
	for true {
		watchForConsoleInput(conn)
	}
}

func getConfig() (string, util.Properties) {
	if len(os.Args) >= 2 {
		username := os.Args[1]
		properties := util.LoadConfig()
		return username, properties
	} else {
		println("You must provide the username as the first parameter ")
		os.Exit(1)
		return "", util.Properties{}
	}
}

func watchForConsoleInput(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)

	for true {
		message, err := reader.ReadString('\n')
		util.CheckForError(err, "Lost console connection")

		message = strings.TrimSpace(message)
		if message != "" {
			command := parseInput(message)

			if command.Command == "" {
				sendCommand("message", message, conn)
			} else {
				switch command.Command {

				case "enter":
					sendCommand("enter", command.Body, conn)

				case "ignore":
					sendCommand("ignore", command.Body, conn)

				case "leave":
					sendCommand("leave", "", conn)

				case "disconnect":
					sendCommand("disconnect", "", conn)

				default:
					fmt.Printf("Unknown command \"%s\"\n", command.Command)
				}
			}
		}
	}
}

func watchForConnectionInput(username string, properties util.Properties, conn net.Conn) {
	reader := bufio.NewReader(conn)

	for true {
		message, err := reader.ReadString('\n')
		util.CheckForError(err, "Lost server connection")
		message = strings.TrimSpace(message)
		if message != "" {
			Command := parseCommand(message)
			switch Command.Command {

			case "ready":
				sendCommand("user", username, conn)

			case "connect":
				fmt.Printf(properties.HasEnteredTheLobbyMessage+"\n", Command.Username)

			case "disconnect":
				fmt.Printf(properties.HasLeftTheLobbyMessage+"\n", Command.Username)

			case "enter":
				fmt.Printf(properties.HasEnteredTheRoomMessage+"\n", Command.Username, Command.Body)

			case "leave":
				fmt.Printf(properties.HasLeftTheRoomMessage+"\n", Command.Username, Command.Body)

			case "message":
				if Command.Username != username {
					fmt.Printf(properties.ReceivedAMessage+"\n", Command.Username, Command.Body)
				}

			case "ignoring":
				fmt.Printf(properties.IgnoringMessage+"\n", Command.Body)
			}
		}
	}
}

func sendCommand(command string, body string, conn net.Conn) {
	message := fmt.Sprintf("/%v %v\n", util.Encode(command), util.Encode(body))
	conn.Write([]byte(message))
}

func parseInput(message string) Command {
	res := standardInputMessageRegex.FindAllStringSubmatch(message, -1)
	if len(res) == 1 {
		return Command{
			Command: res[0][1],
			Body:    res[0][2],
		}
	} else {
		return Command{
			Body: util.Decode(message),
		}
	}
}

func parseCommand(message string) Command {
	res := chatServerResponseRegex.FindAllStringSubmatch(message, -1)
	if len(res) == 1 {
		return Command{
			Command:  util.Decode(res[0][1]),
			Username: util.Decode(res[0][2]),
			Body:     util.Decode(res[0][3]),
		}
	} else {
		return Command{}
	}
}
