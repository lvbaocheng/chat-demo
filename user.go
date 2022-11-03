package main

import "net"

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		conn:   conn,
		C:      make(chan string),
		server: server,
	}
	//启动监听当前User channel消息的goroutine
	go user.ListenMessage()

	return user
}

// 用户上线业务逻辑
func (this *User) Online() {
	//用户上线，将用户加入到OnlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	//广播当前用户上线消息
	this.server.BroadCast(this, "Online")
}

// 用户下线逻辑
func (this *User) Offline() {
	//用户下线，将用户从OnlineMap中去除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	//广播当前用户上线消息
	this.server.BroadCast(this, "Offline")
}

// 给当前user对应的客户端发送消息
func (this *User) sendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 处理用户消息逻辑
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前在线用户数据
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "Online...\n"
			this.sendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else {
		this.server.BroadCast(this, msg)
	}
}

// 监听当前User channel的方法，一旦有消息，就直接发送给对方客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
