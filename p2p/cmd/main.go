package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	p2p "p2p"
	"strings"
	"time"
)

func main() {
	address := flag.String("address", "0.0.0.0:12345", "The tcpv4 on which to listen for tcp connections. (tcp://host:port)")
	flag.Parse()

	var signalc chan os.Signal
	var userinput *bufio.Scanner
	var stdinc chan string

	{
		signalc = make(chan os.Signal)
		signal.Notify(signalc, os.Interrupt)

		stdinc = make(chan string)
		userinput = bufio.NewScanner(os.Stdin)
	}
	g := p2p.NewGater()

	ips := []string{}
	for i := 0; i < 27; i++ {
		ips = append(ips, fmt.Sprintf("172.17.0.1:120%d", i+1))
	}

	if *address == "0.0.0.0:12345" {
		fmt.Println("Default address is being used.")
	}

	port := strings.Split(*address, ":")[1]
	external := fmt.Sprintf("172.17.0.1:%s", port)

	g.Config("tcp", *address, external, ips)
	g.Init()

	go g.Listen()

	<-g.Ready()
	go g.Handle()

	go func() {
		for userinput.Scan() {
			stdinc <- userinput.Text()
		}
	}()

	for run := true; run; {
		select {
		case <-signalc:
			fmt.Println("Interrupting...")
			g.Close()

		case <-g.Done():
			fmt.Println("Closing")
			run = false

		case <-stdinc:
			fmt.Println("Gossiping to peers")
			m := p2p.GossipMessage(*address)
			err := g.Broadcast(m, true)
			if err != nil {
				fmt.Println("Failed to gossip")
			} else {
				fmt.Println("Gossiped successfully")
			}

		case <-time.NewTicker(10 * time.Second).C:
			lastrand := rand.Intn(27)
			if strings.Compare(*address, fmt.Sprintf("0.0.0.0:120%d", lastrand)) == 0 {
				<-time.After(time.Second * 10)
				fmt.Println("Gossiping to peers")
				m := p2p.GossipMessage(*address)
				err := g.Broadcast(m, true)
				if err != nil {
					fmt.Println("Failed to gossip")
				} else {
					fmt.Println("Gossiped successfully")
				}

			}
		}
	}
}
