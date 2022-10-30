# dsys-hw3

Distrubuted Systems Homework #3

```
$ go run . -help
Usage of dsys-hw3:
  -delay int
        Delay in seconds between a message being registered and the same message being sent
  -loss int
        0-100% chance of message (on send) loss
  -name string
        Name of this instance (default "NoName")
  -port string
        Port to connect to (default "50050")
  -start string
        Entrypoint for the application. Either client or server (default "client")
```

## Running it

Given that the architecture is client/server, we need to run one server and *n* clients.
All commands assume that the cwd is the base of the repository.

Run a server:
`go run . -start server -name Chitty`

Then run *n* clients (repeat *n* times):
`go run . -name ClientN`

A client can only start if it successfully connects to a server (i.e. the server must be running first), and all clients close if the server closes.
Use the `-port` flag to specify a port, for the server and clients, if 50050 is not available.

## Using it

The server requires no interaction.

Each client opens as a neat TUI.
Type messages that are previewed at the bottom and press enter to send.
Alternatively, type /help and get a very short list of client commands.
Both /quit and Ctrl-c can be used to quit the client gracefully, letting the server, and hence everyone else, know that you are no longer taking part in the conversation.
The latest message, in terms of wall time, is highlighted in cyan, while the latest message by lamport timestamp is at the bottom of the messages list.

The TUI library used, [termui](https://github.com/gizak/termui), is very representative of the golang ecosystem in that it is not heavily maintained and essential features, like textboxes, have to be found in WIP (read: abandoned) side branches 🙃.
This means: Don't try to send empty messages or the lib will crash. Also, some characters like "[" cannot be typed.

## Inspecting it

Logs for the server is written to stdout **and** to a file `<name>.log`, where `<name>` is the name passed in the `-name` flag.

Logs for the clients are written to a file `<name>.log`, where `<name>` is the name passed in the `-name` flag.

To view them all simultaneously, use a command like `tail -fq *.log`.
Every log for a process is also prefixed with its name, so it is no problem to make sense of interleaved logs.
