package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}

func (s *Server) Handler(conn net.Conn) {
	user := NewUser(conn, s)
	user.Online()

	go func() {
		// 创建4KB的字节切片并从中读取用户信息
		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)

			//下线,异常处理
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Read user message err")
				return
			}

			//广播
			msg := strings.TrimSpace(string(buf[:n]))
			user.DoMessage(msg)
		}
	}()

	select {}

}

func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}

func (s *Server) ListenServerMessage() {
	for {
		serverMessage := <-s.Message
		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.C <- serverMessage
		}
		s.mapLock.Unlock()
	}
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("Failed: net.Listen error")
	}
	defer listener.Close()
	fmt.Printf("Server listening on: %s:%d\n", s.Ip, s.Port)

	go s.ListenServerMessage()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Faild: listener.Accept error, Retrying ...")
		}
		go s.Handler(conn)
	}
}
