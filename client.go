package main

import (
	"context"
	"log"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/JonasUJ/dsys-hw3/chittychat"
	"github.com/JonasUJ/dsys-hw3/lamport"
	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	time     uint64
	pid      uint32 // use pids because we dont care to generate uuids
	stream   chittychat.Chat_ConnectClient
	messages []*chittychat.Message
}

func (client *Client) GetTime() uint64 {
	return client.time
}

func (client *Client) GetPid() uint32 {
	return client.pid
}

func (client *Client) Handle(cmd string) {
	switch cmd {
	case "help":
	case "quit":
		err := client.stream.CloseSend()
		if err != nil {
			// TODO
		}
	default:

	}
}

func (client *Client) Send(msg string) {
	client.time = lamport.LamportSend(client)
	client.stream.Send(lamport.MakeMessage(client, msg))
}

func (client *Client) Recv(msg *chittychat.Message) {
	client.time = lamport.LamportRecv(client, msg)
	client.messages = append(client.messages, msg)
}

func (client *Client) GetRows() []string {
	// It would be very easy to optimise this so so much, but I also just don't care.
	rows := make([]string, len(client.messages))
	sort.SliceStable(client.messages, func(i, j int) bool {
		return lamport.Compare(client.messages[i], client.messages[j]) < 0
	})

	for i, m := range client.messages {
		rows[i] = m.Content
	}

	return rows
}

func client() {
	// Connect to server
	conn, err := grpc.Dial(net.JoinHostPort("localhost", *port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	// Init client
	chatclient := chittychat.NewChatClient(conn)
	stream, err := chatclient.Connect(context.Background())
	if err != nil {
		log.Fatalf("fail to connect: %v", err)
	}

	client := &Client{0, uint32(os.Getpid()), stream, []*chittychat.Message{}}

	// Init ui
	if err := termui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer termui.Close()

	list := widgets.NewList()
	list.Title = "Messages"
	list.TextStyle = termui.NewStyle(termui.ColorYellow)
	list.WrapText = true

	textBox := widgets.NewTextBox()
	textBox.ShowCursor = true
	textBox.Title = "Type message or /help for a list of commands"

	// Helper funcs
	redraw := func() {
		termui.Render(list)
		termui.Render(textBox)
	}

	resize := func(width, height int) {
		list.SetRect(0, 0, width, height-3)
		textBox.SetRect(0, height-3, width, height)
	}

	resize(termui.TerminalDimensions())

	uiEvents := termui.PollEvents()

	// Recv msgs from server
	chMsgs := make(chan *chittychat.Message)
	go func() {
		msg, err := stream.Recv()
		if err != nil {
			return
		}
		chMsgs <- msg
	}()

	for {
		redraw()

		select {
		case e := <-uiEvents:
			switch e.ID {
			case "<Resize>":
				payload := e.Payload.(termui.Resize)
				resize(payload.Width, payload.Height)
			case "<C-c>":
				return
			case "<Left>":
				textBox.MoveCursorLeft()
			case "<Right>":
				textBox.MoveCursorRight()
			case "<Backspace>":
				textBox.Backspace()
			case "<Space>":
				textBox.InsertText(" ")
			case "<Enter>":
				text := textBox.GetText()
				textBox.ClearText()
				if len(text) == 0 {
					continue
				} else if strings.HasPrefix(text, "/") {
					client.Handle(strings.TrimPrefix(text, "/"))
				} else {
					client.Send(text)
				}
			default:
				if termui.ContainsString(termui.PRINTABLE_KEYS, e.ID) {
					textBox.InsertText(e.ID)
				}
			}
		case m := <-chMsgs:
			client.Recv(m)
			list.Rows = client.GetRows()
			list.ScrollBottom()
		}
	}
}
