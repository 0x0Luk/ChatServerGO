package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

type client struct {
	ch       chan<- string
	nickname string
	color    string 
}

var (
	entering      = make(chan client)
	leaving       = make(chan client)
	messages      = make(chan message)
	clients       = make(map[client]bool)
	mutex         sync.Mutex
)

type message struct {
	text   string
	sender client
}

var colors = []string{
	"\033[31m",  
	"\033[1;31m",  
	"\033[91m",  
	
	"\033[32m",  
	"\033[1;32m",  
	"\033[92m",  
	
	"\033[33m",  
	"\033[1;33m", 
	"\033[93m",  
	
	"\033[34m",  
	"\033[1;34m", 
	"\033[94m",  
	
	"\033[35m",  
	"\033[1;35m",  
	"\033[95m",  
	
	"\033[36m",  
	"\033[1;36m",  
	"\033[96m",  
}


func colorize(colorCode, text string) string {
	return fmt.Sprintf("%s%s\033[0m", colorCode, text)
}

func randomColor() string {
	rand.Seed(time.Now().UnixNano())
	return colors[rand.Intn(len(colors))]
}

func broadcaster() {
	for {
		select {
		case msg := <-messages:
			mutex.Lock()
			for cli := range clients {
				if cli != msg.sender { 
					cli.ch <- msg.text
				}
			}
			mutex.Unlock()

		case cli := <-entering:
			mutex.Lock()
			clients[cli] = true
			mutex.Unlock()

		case cli := <-leaving:
			mutex.Lock()
			delete(clients, cli)
			close(cli.ch)
			mutex.Unlock()
		}
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	var nickname string

	go clientWriter(conn, ch)

	input := bufio.NewScanner(conn)
	if input.Scan() {
		nickname = input.Text()
	}

	userColor := randomColor()

	newClient := client{ch: ch, nickname: nickname, color: userColor}
	welcomeMsg := colorize(userColor, fmt.Sprintf("Usuário @%s acabou de entrar", nickname))
	messages <- message{text: welcomeMsg, sender: newClient}
	entering <- newClient

	for input.Scan() {
		text := input.Text()

		if strings.HasPrefix(text, "\\changenick") {
			parts := strings.Split(text, " ")
			if len(parts) == 2 {
				oldNick := nickname
				nickname = parts[1]
				newClient.nickname = nickname
				changeNickMsg := colorize(userColor, fmt.Sprintf("Usuário @%s agora é @%s", oldNick, nickname))
				messages <- message{text: changeNickMsg, sender: newClient}
			}

		} else if strings.HasPrefix(text, "\\msg") {
			parts := strings.Split(text, " ")
			if len(parts) >= 3 {
				target := strings.TrimPrefix(parts[1], "@")
				msg := strings.Join(parts[2:], " ")

				mutex.Lock()
				for cli := range clients {
					if cli.nickname == target {
						privateMsg := colorize("\033[35m", fmt.Sprintf("%s disse em privado: %s", nickname, msg))
						cli.ch <- privateMsg
						break
					}
				}
				mutex.Unlock()
			}

		} else {
			publicMsg := colorize(userColor, fmt.Sprintf("@%s disse: %s", nickname, text))
			messages <- message{text: publicMsg, sender: newClient}
		}
	}

	leaving <- newClient
	goodbyeMsg := colorize(userColor, fmt.Sprintf("Usuário @%s saiu", nickname))
	messages <- message{text: goodbyeMsg, sender: newClient}
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func main() {
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Servidor rodando na porta 3000")

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}
