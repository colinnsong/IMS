package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	IP   string
	Port int

	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// 执行业务的接口
func (server *Server) Handler(conn net.Conn) {
	fmt.Println("客户端连接成功")

	// 创建User实例并将其加入到OnlineMap中
	user := NewUser(conn, server)

	// 创建记录当前用户是否活跃的channel，用于和负责处理客户端消息的go程通信
	isActive := make(chan bool)

	// 用户上线
	user.Online()

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				// 用户下线
				user.Offline()
				fmt.Println("客户端断开连接")
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("conn.Read err:", err)
				return
			}
			// 提取用户消息并去除'\n'
			msg := string(buf[:n-1])
			// 处理用户消息
			user.DoMessage(msg)
			// 用户活跃
			isActive <- true
		}
	}()

	// 阻塞当前go程
	for {
		select {
		case <-isActive:
			// 用户活跃，可以不做操作，自动执行下一个case重置定时器
		case <-time.After(10 * time.Second):
			// 用户登录超时，强制下线用户
			user.conn.Write([]byte("你被踢了\n"))
			close(user.channel)
			conn.Close()

			// 退出函数，结束当前go程
			return
		}
	}

}

// 广播用户上线消息的接口
func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	// 将消息发送到消息广播的channel中
	server.Message <- sendMsg
}

// 监听Message中是否有消息
func (server *Server) ListenMessage() {
	for {
		msg := <-server.Message

		// 如果有消息，就将消息发送给OnlineMap中的所有用户
		server.mapLock.Lock()
		for _, user := range server.OnlineMap {
			user.channel <- msg
		}
		server.mapLock.Unlock()
	}
}

// 启动服务器的接口
func (server *Server) Start() {
	// 创建socket并监听
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.IP, server.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	defer listener.Close()

	// 启动go程监听Message中是否有消息
	go server.ListenMessage()

	// 循环等待客户端连接并执行相关业务
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}

		// 创建go程执行相关业务
		go server.Handler(conn)
	}
}

// 创建Server实例的接口
func NewServer(ip string, port int) *Server {
	return &Server{
		IP:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}
