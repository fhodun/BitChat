package main

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/fhodun/BitChat/server/internal/endpoint/json"
	"github.com/fhodun/BitChat/server/internal/util"
)

const lobby = "lobby"

func main() {
	properties := util.LoadConfig()
	psock, err := net.Listen("tcp", ":"+properties.Port)
	util.CheckForError(err, "Can't create server")

	fmt.Printf("Chat server started on port %v...\n", properties.Port)

	go json.Start()

	for {
		conn, err := psock.Accept()
		util.CheckForError(err, "Can't accept connections")

		client := util.Client{Connection: conn, Room: lobby, Properties: properties}
		client.Register()

		channel := make(chan string)
		go waitForInput(channel, &client)
		go handleInput(channel, &client, properties)

		util.SendClientMessage("ready", properties.Port, &client, true, properties)
	}
}

func waitForInput(out chan string, client *util.Client) {
	defer close(out)

	reader := bufio.NewReader(client.Connection)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			client.Close(true)
			return
		}
		out <- string(line)
	}
}

func handleInput(in <-chan string, client *util.Client, props util.Properties) {

	for {
		message := <-in
		if message != "" {
			message = strings.TrimSpace(message)
			action, body := getAction(message)

			if action != "" {
				switch action {

				case "message":
					util.SendClientMessage("message", body, client, false, props)

				case "user":
					client.Username = body
					util.SendClientMessage("connect", "", client, false, props)

				case "disconnect":
					client.Close(false)

				case "ignore":
					client.Ignore(body)
					util.SendClientMessage("ignoring", body, client, false, props)

				case "enter":
					if body != "" {
						client.Room = body
						util.SendClientMessage("enter", body, client, false, props)
					}

				case "leave":
					if client.Room != lobby {
						util.SendClientMessage("leave", client.Room, client, false, props)
						client.Room = lobby
					}

				default:
					util.SendClientMessage("unrecognized", action, client, true, props)
				}
			}
		}
	}
}

func getAction(message string) (string, string) {
	actionRegex, _ := regexp.Compile(`^\/([^\s]*)\s*(.*)$`)
	res := actionRegex.FindAllStringSubmatch(message, -1)
	if len(res) == 1 {
		return res[0][1], res[0][2]
	}
	return "", ""
}
