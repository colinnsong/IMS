package main

import "net"

type User struct {
	Name    string
	Addr    string
	channel chan string
	conn    net.Conn
	server  *Server
}

// 创建User实例的接口
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:    userAddr,
		Addr:    userAddr,
		channel: make(chan string),
		conn:    conn,
		// 将server对象和当前的user对象绑定
		server: server,
	}

	// 启动go程监听channel中是否有消息
	go user.ListenMessage()

	return user
}

// 监听channel中是否有消息
func (user *User) ListenMessage() {
	for {
		msg := <-user.channel
		user.conn.Write([]byte(msg + "\n"))
	}
}

// 用户上线业务处理的接口
func (user *User) Online() {
	// 记录当前用户连接
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()
	// 广播用户上线的消息
	user.server.BroadCast(user, "已上线")
}

// 用户下线业务处理接口
func (user *User) Offline() {
	// 删除当前用户连接
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()
	// 广播用户下线的消息
	user.server.BroadCast(user, "已下线")
}

// 用户处理消息的接口
func (user *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户
		user.server.mapLock.Lock()
		for _, onlineUser := range user.server.OnlineMap {
			onlineMsg := "[" + onlineUser.Addr + "]" + onlineUser.Name + ":" + "在线...\n"
			user.conn.Write([]byte(onlineMsg))
		}
		user.server.mapLock.Unlock()
	} else if msg[:7] == "rename|" {
		// 修改当前用户的名称
		newName := msg[7:]
		user.server.mapLock.Lock()
		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.conn.Write([]byte("当前用户名已被使用\n"))
		} else {
			user.server.OnlineMap[newName] = user
			delete(user.server.OnlineMap, user.Name)
			user.Name = newName
			user.conn.Write([]byte("已成功修改用户名为：" + newName + "\n"))
		}
		user.server.mapLock.Unlock()
	} else {
		user.server.BroadCast(user, msg)
	}

}
