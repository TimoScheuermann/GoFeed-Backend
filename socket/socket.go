package socket

import (
	"fmt"

	socketio "github.com/googollee/go-socket.io"
)

var SocketServer *socketio.Server

func init() {
	fmt.Println("Init Socket")
	SocketServer = socketio.NewServer(nil)

	SocketServer.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		fmt.Println("connected: ", s.ID())
		s.Join("clients")
		return nil
	})

	go SocketServer.Serve()
	defer SocketServer.Close()
}
