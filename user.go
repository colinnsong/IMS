package main

import "net"

type User struct {
	Name    string
	Addr    string
	channel chan string
	conn    net.Conn
}

// 创建User实例的接口
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:    userAddr,
		Addr:    userAddr,
		channel: make(chan string),
		conn:    conn,
	}

	// 启动go程监听channel中是否有消息
	go user.ListenMessage()

	return user
}

// 监听channel中是否有消息的接口
func (user *User) ListenMessage() {
	for {
		msg := <-user.channel
		user.conn.Write([]byte(msg + "\n"))
	}
}
