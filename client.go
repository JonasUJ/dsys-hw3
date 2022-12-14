package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/JonasUJ/dsys-hw3/chittychat"
	"github.com/JonasUJ/dsys-hw3/lamport"
	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Event struct {
	ID      string
	Message *chittychat.Message
}

type Client struct {
	time     uint64
	pid      uint32 // use pids because we dont care to generate uuids
	stream   chittychat.Chat_ConnectClient
	messages []*chittychat.Message
	events   chan *Event
}

func NewClient(stream chittychat.Chat_ConnectClient) *Client {
	client := &Client{
		pid:      uint32(os.Getpid()),
		stream:   stream,
		messages: []*chittychat.Message{},
		events:   make(chan *Event, 1),
	}

	// Recv msgs from server in background
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				l.Println("lost connection to server")
				// Quit because there's no option to reestablish connection besides running the
				// process again.
				client.events <- &Event{"quit", nil}
				return
			}

			// Ignore messages from ourself
			if msg.Pid != client.pid {
				client.events <- &Event{"msg", msg}
			}
		}
	}()

	return client
}

// -- start Lamport interface --

func (client *Client) GetTime() uint64 {
	return client.time
}

func (client *Client) GetPid() uint32 {
	return client.pid
}

// -- end Lamport interface --

// Handle user commands typed in the text box.
// We only really need these because the requirements say clients should be able to quit.
func (client *Client) Handle(cmd string) {
	parts := strings.Split(cmd, " ")
	switch parts[0] {
	case "help":
		client.Log(`[Type messages and press enter to send.](fg:magenta)
Available commands:
/[quit](fg:green) - Gracefully exits the chatroom
/[loss](fg:green) [<percent>](fg:blue) - Set message loss percent
/[delay](fg:green) [<seconds>](fg:blue) - Set message delay in seconds
/[clear](fg:green) - Forgets all messages
/[help](fg:green) - Displays this message`)
	case "loss":
		if len(parts) < 2 {
			client.Log(fmt.Sprintf("Current loss is [%d%%](fg:green)", *loss))
			return
		}

		amount, err := strconv.Atoi(parts[1])
		if err != nil {
			client.Log(fmt.Sprintf("['%s' is not a valid integer](fg:red)", parts[1]))
			return
		}

		client.Log(fmt.Sprintf("Changed loss from [%d%%](fg:red) to [%d%%](fg:green)", *loss, amount))
		*loss = amount
	case "delay":
		if len(parts) < 2 {
			client.Log(fmt.Sprintf("Current delay is [%ds](fg:green)", *delay))
			return
		}

		amount, err := strconv.Atoi(parts[1])
		if err != nil {
			client.Log(fmt.Sprintf("['%s' is not a valid integer](fg:red)", parts[1]))
			return
		}

		client.Log(fmt.Sprintf("Changed delay from [%ds](fg:red) to [%ds](fg:green)", *delay, amount))
		*delay = amount
	case "clear":
		client.messages = []*chittychat.Message{}
		client.events <- &Event{"clear", nil}
	case "quit":
		err := client.stream.CloseSend()
		if err != nil {
			l.Fatalf("fail to close stream: %v", err)
		}
		client.events <- &Event{"quit", nil}
	default:
		client.Log(fmt.Sprintf("[Unknown command '%s'](fg:red)", cmd))
	}
}

// Send handler for messages. Makes sure we remember to increment time.
func (client *Client) Send(msg string) {
	time := lamport.LamportSend(client)
	l.Printf("ticking client time (%d -> %d)", client.time, time)
	client.time = time

	// Cool colors for our own message :)
	ownMsg := fmt.Sprintf("[%s>](fg:green) %s", *name, msg)
	sendMsg := fmt.Sprintf("[%s>](fg:blue) %s", *name, msg)

	client.Log(ownMsg)

	// Check if this message was randomly "lost"
	if !Lost() {
		msg := lamport.MakeMessage(client, sendMsg)
		Delay(func() {
			client.stream.Send(msg)
		})
	}
}

// Recv handler for messages. Makes sure we remember to increment time.
func (client *Client) Recv(msg *chittychat.Message) {
	time := lamport.LamportRecv(client, msg)
	l.Printf("ticking client time (%d -> %d)", client.time, time)
	client.time = time
	client.messages = append(client.messages, msg)
}

// Log a message locally (without sending it to the server)
func (client *Client) Log(msg string) {
	client.events <- &Event{"msg", lamport.MakeMessage(client, msg)}
}

// Get all messages sorted by time in ascending order
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
		l.Fatalf("fail to dial: %v", err)
	}

	// Init client
	chatclient := chittychat.NewChatClient(conn)
	stream, err := chatclient.Connect(context.Background())
	if err != nil {
		l.Fatalf("fail to connect: %v", err)
	}

	client := NewClient(stream)

	// Tell server our name so that it can tell everyone else
	client.time = lamport.LamportSend(client)
	client.stream.Send(lamport.MakeMessage(client, *name))

	// Init ui
	if err := termui.Init(); err != nil {
		l.Fatalf("failed to initialize termui: %v", err)
	}
	defer termui.Close()

	list := widgets.NewList()
	list.Title = "Messages"
	list.TextStyle = termui.NewStyle(termui.ColorYellow)
	list.SelectedRowStyle = termui.NewStyle(termui.ColorCyan)
	list.WrapText = true

	textBox := widgets.NewTextBox()
	textBox.ShowCursor = true
	textBox.Title = "Type message or /help for a list of commands"

	// Helper ui funcs
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

	// Main loop. Handles user input and displaying new messages from the server.
	for {
		redraw()

		select {
		case e := <-uiEvents:
			switch e.ID {
			case "<Resize>":
				payload := e.Payload.(termui.Resize)
				resize(payload.Width, payload.Height)
			case "<C-c>":
				client.Handle("quit")
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
		case e := <-client.events:
			l.Printf("client event '%s'", e)
			switch e.ID {
			case "msg":
				client.Recv(e.Message)
				list.Rows = client.GetRows()
				list.ScrollTop()
				list.ScrollAmount(slices.Index(client.messages, e.Message))
			case "clear":
				list.Rows = client.GetRows()
			case "quit":
				return
			}
		}
	}
}
