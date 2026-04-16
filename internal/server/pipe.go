package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/google/uuid"
	"github.com/scotmcc/pigo/internal/commands"
)

// Pipe manages the persistent TCP pipe connection.
// One pipe per client. Commands are routed by GUID.
type Pipe struct {
	listener net.Listener
	addr     string

	// mu protects the pipeConn — the single registered pipe connection.
	mu       sync.Mutex
	pipeConn net.Conn
	encoder  *json.Encoder
}

// NewPipe creates a TCP pipe server on the given address.
func NewPipe(host string, port int) *Pipe {
	return &Pipe{
		addr: fmt.Sprintf("%s:%d", host, port),
	}
}

// Start begins listening for pipe connections. Blocks until closed.
func (p *Pipe) Start() error {
	var err error
	p.listener, err = net.Listen("tcp", p.addr)
	if err != nil {
		return fmt.Errorf("pipe listen on %s: %w", p.addr, err)
	}
	fmt.Printf("pigo pipe listening on %s\n", p.addr)

	for {
		conn, err := p.listener.Accept()
		if err != nil {
			return nil // listener closed (normal shutdown)
		}
		go p.handleConnection(conn)
	}
}

// Close shuts down the pipe listener.
func (p *Pipe) Close() error {
	if p.listener != nil {
		return p.listener.Close()
	}
	return nil
}

// handleConnection reads newline-delimited JSON from a connection.
// If the first message is a register_pipe request, this becomes the pipe.
// Otherwise each message is dispatched as a command.
func (p *Pipe) handleConnection(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	// Increase buffer for large payloads.
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			p.sendToConn(conn, Response{Error: "invalid JSON: " + err.Error()})
			continue
		}

		if req.Type == "register_pipe" {
			p.registerPipe(conn)
			continue
		}

		if req.Command == "" {
			p.sendToConn(conn, Response{GUID: req.GUID, Error: "command is required"})
			continue
		}

		p.dispatchCommand(conn, req)
	}

	// Connection closed — if this was the pipe, clear it.
	p.mu.Lock()
	if p.pipeConn == conn {
		p.pipeConn = nil
		p.encoder = nil
		fmt.Println("pipe disconnected")
	}
	p.mu.Unlock()
}

// registerPipe marks this connection as the persistent pipe.
func (p *Pipe) registerPipe(conn net.Conn) {
	p.mu.Lock()
	p.pipeConn = conn
	p.encoder = json.NewEncoder(conn)
	p.mu.Unlock()

	fmt.Printf("pipe registered from %s\n", conn.RemoteAddr())
	p.sendToConn(conn, Response{Success: true, Status: "ok", Message: "pipe registered"})
}

// dispatchCommand handles a command request from the pipe.
// Sync commands respond on the same connection.
// Async commands send an immediate ack, then stream updates through the pipe.
func (p *Pipe) dispatchCommand(conn net.Conn, req Request) {
	guid := req.GUID
	if guid == "" {
		guid = uuid.New().String()
	}

	if commands.IsAsync(req.Command) {
		// Async: ack immediately, run in background.
		p.sendToConn(conn, Response{GUID: guid, Status: "accepted", Message: "running in background"})

		go func() {
			send := p.makeSendFunc(guid)
			data, err := commands.Dispatch(req.Command, req.Args, send)
			if err != nil {
				p.sendToPipe(Response{GUID: guid, Status: "error", Error: err.Error()})
				return
			}
			p.sendToPipe(Response{GUID: guid, Status: "done", Success: true, Data: data})
		}()
	} else {
		// Sync: run inline, respond on same connection.
		send := p.makeSendFunc(guid)
		data, err := commands.Dispatch(req.Command, req.Args, send)
		if err != nil {
			p.sendToConn(conn, Response{GUID: guid, Success: false, Error: err.Error()})
			return
		}
		p.sendToConn(conn, Response{GUID: guid, Success: true, Data: data})
	}
}

// makeSendFunc creates a SendFunc that routes updates through the pipe by GUID.
func (p *Pipe) makeSendFunc(guid string) commands.SendFunc {
	return func(status, message string, data any) error {
		return p.sendToPipe(Response{
			GUID:    guid,
			Status:  status,
			Message: message,
			Data:    data,
		})
	}
}

// sendToPipe sends a response through the registered pipe connection.
func (p *Pipe) sendToPipe(resp Response) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.encoder == nil {
		return fmt.Errorf("no pipe connected")
	}

	return p.encoder.Encode(resp)
}

// sendToConn sends a newline-delimited JSON response to a specific connection.
func (p *Pipe) sendToConn(conn net.Conn, resp Response) {
	json.NewEncoder(conn).Encode(resp)
}
