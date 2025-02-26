package main

import (
	"flag"
	"fmt"
	"io"
	"net"
)

type Client struct {
	ServerIP   string
	ServerPort int
	conn       net.Conn
	flag       int
}

func (client *Client) Menu() bool {
	var flag int
	fmt.Println("1. 公聊模式")
	fmt.Println("2. 私聊模式")
	fmt.Println("3. 更新用户名")
	fmt.Println("0. 退出")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("请输入合法范围内的数字")
		return false
	}
}

// 更新用户名
func (client *Client) Rename() bool {
	var userName string
	fmt.Println("请输入用户名:")
	fmt.Scanln(&userName)
	sendMsg := "rename|" + userName + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

// 公聊模式
func (client *Client) PublicChat() {
	var chatMsg string
	fmt.Println("请输入聊天内容, exit退出")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				break
			}
		} else {
			fmt.Println("消息不能为空")
		}
		chatMsg = ""
		fmt.Println("请输入聊天内容, exit退出")
		fmt.Scanln(&chatMsg)
	}
}

// 查询在线用户
func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return
	}
}

// 私聊模式
func (client *Client) PrivateChat() {
	client.SelectUsers()

	var remoteName string
	fmt.Println("请输入聊天对象用户名, exit退出")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		var chatMsg string
		fmt.Println("请输入消息内容, exit退出")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn.Write err:", err)
					break
				}
			} else {
				fmt.Println("消息不能为空")
			}
			chatMsg = ""
			fmt.Println("请输入消息内容, exit退出")
			fmt.Scanln(&chatMsg)
		}
		fmt.Println("请输入聊天对象用户名, exit退出")
		fmt.Scanln(&remoteName)
	}
}

// 执行客户端业务
func (client *Client) Run() {
	for client.Menu() {
		switch client.flag {
		case 1:
			client.PublicChat()
		case 2:
			client.PrivateChat()
		case 3:
			client.Rename()
		case 0:
			return
		}
	}
}

// 监听客户端收到的消息
func (client *Client) DealResponse() {
	buf := make([]byte, 4096)
	for {
		n, err := client.conn.Read(buf)
		if n == 0 {
			// 服务端关闭
			fmt.Println("服务器断开连接")
			return
		}

		if err != nil && err != io.EOF {
			fmt.Println("conn.Read err:", err)
			return
		}

		fmt.Println(string(buf[:n-1]))
	}
}

func NewClient(serverIP string, serverPort int) *Client {
	client := &Client{
		ServerIP:   serverIP,
		ServerPort: serverPort,
		flag:       999,
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIP, serverPort))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return nil
	}
	client.conn = conn
	return client
}

func main() {
	// 定义命令行参数
	serverIP := flag.String("ip", "127.0.0.1", "服务器IP地址")
	serverPort := flag.Int("port", 8888, "服务器端口")

	// 解析命令行参数
	flag.Parse()

	client := NewClient(*serverIP, *serverPort)
	if client == nil {
		fmt.Println("连接服务器失败")
		return
	}
	fmt.Println("连接服务器成功")

	// 启动监听客户端收到的消息的go程
	go client.DealResponse()

	// 启动客户端业务
	client.Run()

	// 客户端退出之前关闭和服务器的连接
	defer client.conn.Close()
}
