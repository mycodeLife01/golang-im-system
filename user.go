package main

import (
	"fmt"
	"net"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	addr := conn.RemoteAddr().String()
	user := &User{
		Name:   addr,
		Addr:   addr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	go user.ListenMessage()
	return user
}

func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}

// 上线业务
func (u *User) Online() {
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	u.server.BroadCast(u, "上线了")
}

// 下线业务
func (u *User) Offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	u.server.BroadCast(u, "下线了")
}

// 发送消息
func (u *User) DoMessage(msg string) {
	fmt.Println("DoMessage里的msg是", msg)
	if msg == "who" {
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := "[" + u.Addr + "]:" + u.Name + "在线 ...\n"
			user.sendMsg(onlineMsg)
		}
		u.server.mapLock.Unlock()
	} else {
		u.server.BroadCast(u, msg)
	}

}

// 给当前用户发送消息
func (u *User) sendMsg(msg string) {
	u.conn.Write([]byte(msg + "\n"))
}
