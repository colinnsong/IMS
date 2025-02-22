package main

import (
	"fmt"
	"net"
)

type Server struct {
	IP   string
	Port int
}

// Server类执行业务的成员方法
func (server *Server) Handler(conn net.Conn) {
	fmt.Println("客户端连接成功")
}

// 创建Server实例的接口
func NewServer(ip string, port int) *Server {
	return &Server{
		IP:   ip,
		Port: port,
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
