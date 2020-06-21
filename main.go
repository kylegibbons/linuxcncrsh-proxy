package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	sendToLinuxCNC := make(chan string, 100)

	go LinuxCNCConnector(sendToLinuxCNC)

	// Listen for incoming connections.
	l, err := net.Listen("tcp", ":3333")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening")
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		log.Printf("Client Connected: %s", conn.RemoteAddr().String())
		// Handle connections in a new goroutine.
		go handleRequest(conn, sendToLinuxCNC)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn, outgoing chan string) {
	defer conn.Close()

	connbuf := bufio.NewReader(conn)

	for {
		str, err := connbuf.ReadString('\n')
		if len(str) > 0 {
			log.Printf("Receved from client: %s", str)
			outgoing <- str
		}
		if err != nil {
			break
		}

		time.Sleep(100)
	}
}

// LinuxCNCConnector is
func LinuxCNCConnector(incoming chan string) {
	linuxCNCchan := make(chan string, 100)

	conn, err := net.Dial("tcp", "192.168.9.127:5007")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer conn.Close()

	go LinuxCNCReceiver(conn, linuxCNCchan)

	conn.Write([]byte("hello EMC user-at-telnet 1.0\n"))
	//conn.Write([]byte("get kinematics_type\n"))

	for {
		select {
		case msg := <-linuxCNCchan:
			log.Printf("Message from LinuxCNC: %s", msg)

			conn2, err := net.Dial("tcp", "192.168.9.127:5008")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			conn2.Write([]byte(msg))
			conn2.Close()
		case msg := <-incoming:
			log.Printf("Sending to LinuxCNC message: %s", msg)
			_, err := conn.Write([]byte(msg))

			if err != nil {
				log.Printf("Error writing to LinuxCNC: %v", err)
			}

		default:
			//fmt.Println("no message received")
		}
	}
}

//LinuxCNCReceiver is
func LinuxCNCReceiver(conn net.Conn, messages chan string) {

	connbuf := bufio.NewReader(conn)
	for {
		str, err := connbuf.ReadString('\n')
		if len(str) > 0 {
			//log.Printf("Receiver: %s", str)
			messages <- str
		}
		if err != nil {
			log.Printf("%v", err)
			break
		}
	}
}
