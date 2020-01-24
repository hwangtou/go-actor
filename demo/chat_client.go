package main

import (
	"bufio"
	"fmt"
	"github.com/go-actor/demo/chat"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	u := url.URL{
		Scheme:"ws",
		Host:"127.0.0.1:20000",
		Path:"/chat",
	}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("client dial error,", err)
	}
	defer conn.Close()
	go func() {
		for {
			msg := &chat.Response{}
			if err := conn.ReadJSON(msg); err != nil {
				log.Fatal("client read error,", err)
			}
			log.Printf("client receive message, %v\n", msg)
		}
	}()
	reader := bufio.NewReader(os.Stdin)
	for {
		<-time.After(time.Second)
		fmt.Printf("Enter command: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		cmd := strings.Split(text, ":")
		log.Println(cmd)
		content := map[string]interface{}{
			"cmd": cmd[0],
		}
		if len(cmd) > 1 {
			content["data"] = cmd[1]
		}
		if err := conn.WriteJSON(content); err != nil {
			log.Println("client send error,", err)
			break
		}
		log.Println("sending,", content)
	}
}
