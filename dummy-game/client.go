// Dummy Game Client — CLI клиент для тестирования оркестрации.
//
// Сценарий использования:
//   1. Discovery — узнаём доступные серверы через gateway
//   2. Queue — если серверы заняты, встаём в очередь и ждём
//   3. Connect — подключаемся по TCP к игровому серверу
//   4. Gameplay — отправляем команды (ping, chat, list)
//
// Использование:
//   go run client.go -gateway http://localhost:8080 -game-id 1 -player-id alice
//   go run client.go -direct localhost:7777 -player-id bob
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════════
// Gateway API Types
// ═══════════════════════════════════════════════════════════════════════════════

type DiscoveryResult struct {
	Status  string           `json:"status"`
	Message string           `json:"message,omitempty"`
	Servers []ServerEndpoint `json:"servers"`
}

type ServerEndpoint struct {
	InstanceID  string `json:"instance_id"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	PlayerCount int    `json:"player_count"`
	MaxPlayers  int    `json:"max_players"`
}

type QueueStatus struct {
	Status             string         `json:"status"`
	Position           int            `json:"position"`
	TotalInQueue       int            `json:"total_in_queue"`
	EstimatedWaitSec   int            `json:"estimated_wait_seconds"`
	ReservedEndpoint   *ServerEndpoint `json:"reserved_endpoint,omitempty"`
	ReservedUntilUnix  int64          `json:"reserved_until_unix,omitempty"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// Game Protocol Types
// ═══════════════════════════════════════════════════════════════════════════════

type GameCmd struct {
	Cmd     string `json:"cmd"`
	Name    string `json:"name,omitempty"`
	Message string `json:"message,omitempty"`
}

type GameResp struct {
	Cmd      string   `json:"cmd"`
	OK       bool     `json:"ok"`
	Error    string   `json:"error,omitempty"`
	Time     string   `json:"time,omitempty"`
	Players  []string `json:"players,omitempty"`
	Player   string   `json:"player,omitempty"`
	Message  string   `json:"message,omitempty"`
	From     string   `json:"from,omitempty"`
	Count    int      `json:"count,omitempty"`
	Max      int      `json:"max,omitempty"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// Client
// ═══════════════════════════════════════════════════════════════════════════════

type Client struct {
	gateway    string
	gameID     int64
	playerID   string
	httpClient *http.Client
}

func NewClient(gateway string, gameID int64, playerID string) *Client {
	return &Client{
		gateway:    strings.TrimRight(gateway, "/"),
		gameID:     gameID,
		playerID:   playerID,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// ─── Discovery ────────────────────────────────────────────────────────────────

func (c *Client) Discover() (*DiscoveryResult, error) {
	url := fmt.Sprintf("%s/api/v1/games/%d/discover", c.gateway, c.gameID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("discovery request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discovery failed: %s: %s", resp.Status, string(body))
	}

	var result DiscoveryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode discovery: %w", err)
	}
	return &result, nil
}

// ─── Queue ────────────────────────────────────────────────────────────────────

func (c *Client) QueueJoin() (*QueueStatus, error) {
	url := fmt.Sprintf("%s/api/v1/games/%d/queue/join", c.gateway, c.gameID)
	body, _ := json.Marshal(map[string]string{"player_id": c.playerID})
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("queue join: %w", err)
	}
	defer resp.Body.Close()

	var qs QueueStatus
	if err := json.NewDecoder(resp.Body).Decode(&qs); err != nil {
		return nil, fmt.Errorf("decode queue join: %w", err)
	}
	return &qs, nil
}

func (c *Client) QueueHeartbeat() (*QueueStatus, error) {
	url := fmt.Sprintf("%s/api/v1/games/%d/queue/heartbeat", c.gateway, c.gameID)
	body, _ := json.Marshal(map[string]string{"player_id": c.playerID})
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("queue heartbeat: %w", err)
	}
	defer resp.Body.Close()

	var qs QueueStatus
	if err := json.NewDecoder(resp.Body).Decode(&qs); err != nil {
		return nil, fmt.Errorf("decode heartbeat: %w", err)
	}
	return &qs, nil
}

func (c *Client) QueueLeave() error {
	url := fmt.Sprintf("%s/api/v1/games/%d/queue/leave?player_id=%s", c.gateway, c.gameID, c.playerID)
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("queue leave: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

// ═══════════════════════════════════════════════════════════════════════════════
// Connection Flow
// ═══════════════════════════════════════════════════════════════════════════════

func (c *Client) Connect() (net.Conn, *ServerEndpoint, error) {
	fmt.Printf("[%s] Discovering servers for game %d...\n", c.playerID, c.gameID)

	for {
		disc, err := c.Discover()
		if err != nil {
			return nil, nil, err
		}

		switch disc.Status {
		case "ready", "DISCOVERY_STATUS_READY":
			if len(disc.Servers) == 0 {
				return nil, nil, fmt.Errorf("no servers available")
			}
			server := disc.Servers[0]
			fmt.Printf("[%s] Status: READY → connecting to %s:%d\n", c.playerID, server.Address, server.Port)
			return c.dialGame(&server)

		case "reserved", "DISCOVERY_STATUS_RESERVED":
			// reserved для этого player_id через queue
			fmt.Printf("[%s] Status: RESERVED → connecting\n", c.playerID)
			if len(disc.Servers) > 0 {
				return c.dialGame(&disc.Servers[0])
			}
			return nil, nil, fmt.Errorf("reserved but no endpoint")

		case "starting", "DISCOVERY_STATUS_STARTING":
			fmt.Printf("[%s] Status: STARTING → waiting 3s...\n", c.playerID)
			time.Sleep(3 * time.Second)
			continue

		case "queue", "DISCOVERY_STATUS_QUEUE":
			fmt.Printf("[%s] Status: QUEUE → joining queue...\n", c.playerID)
			return c.waitInQueue()

		case "capacity_reached", "DISCOVERY_STATUS_CAPACITY_REACHED":
			return nil, nil, fmt.Errorf("capacity reached: %s", disc.Message)

		case "unavailable", "DISCOVERY_STATUS_UNAVAILABLE":
			return nil, nil, fmt.Errorf("unavailable: %s", disc.Message)

		default:
			fmt.Printf("[%s] Status: %s → waiting 3s...\n", c.playerID, disc.Status)
			time.Sleep(3 * time.Second)
			continue
		}
	}
}

func (c *Client) waitInQueue() (net.Conn, *ServerEndpoint, error) {
	qs, err := c.QueueJoin()
	if err != nil {
		return nil, nil, fmt.Errorf("join queue: %w", err)
	}
	fmt.Printf("[%s] Joined queue: position %d/%d, est. wait %ds\n",
		c.playerID, qs.Position, qs.TotalInQueue, qs.EstimatedWaitSec)

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			qs, err := c.QueueHeartbeat()
			if err != nil {
				return nil, nil, fmt.Errorf("queue heartbeat: %w", err)
			}

			switch qs.Status {
			case "waiting", "QUEUE_STATUS_WAITING":
				fmt.Printf("[%s] Queue: position %d/%d, est. wait %ds\n",
					c.playerID, qs.Position, qs.TotalInQueue, qs.EstimatedWaitSec)

			case "reserved", "QUEUE_STATUS_RESERVED":
				if qs.ReservedEndpoint != nil {
					fmt.Printf("[%s] Queue: RESERVED → connecting to %s:%d\n",
						c.playerID, qs.ReservedEndpoint.Address, qs.ReservedEndpoint.Port)
					return c.dialGame(qs.ReservedEndpoint)
				}
				return nil, nil, fmt.Errorf("reserved but no endpoint")

			case "expired", "QUEUE_STATUS_EXPIRED":
				return nil, nil, fmt.Errorf("queue expired")

			default:
				fmt.Printf("[%s] Queue: %s\n", c.playerID, qs.Status)
			}
		}
	}
}

func (c *Client) dialGame(server *ServerEndpoint) (net.Conn, *ServerEndpoint, error) {
	addr := fmt.Sprintf("%s:%d", server.Address, server.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("dial %s: %w", addr, err)
	}
	return conn, server, nil
}

// ═══════════════════════════════════════════════════════════════════════════════
// Game Session
// ═══════════════════════════════════════════════════════════════════════════════

func (c *Client) Play(conn net.Conn) error {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(reader)

	// Join the game
	if err := encoder.Encode(GameCmd{Cmd: "join", Name: c.playerID}); err != nil {
		return fmt.Errorf("send join: %w", err)
	}
	var joinResp GameResp
	if err := decoder.Decode(&joinResp); err != nil {
		return fmt.Errorf("read join: %w", err)
	}
	if !joinResp.OK {
		return fmt.Errorf("join failed: %s", joinResp.Error)
	}
	fmt.Printf("[%s] Joined game! Players: %d/%d\n", c.playerID, joinResp.Count, joinResp.Max)

	// Read server messages in background
	msgCh := make(chan GameResp, 16)
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		for {
			var resp GameResp
			if err := decoder.Decode(&resp); err != nil {
				if err != io.EOF {
					fmt.Printf("[%s] Read error: %v\n", c.playerID, err)
				}
				return
			}
			select {
			case msgCh <- resp:
			case <-doneCh:
				return
			}
		}
	}()

	// Interactive CLI
	fmt.Println("Commands: ping, list, chat <msg>, leave, quit")
	stdin := bufio.NewReader(os.Stdin)

	for {
		select {
		case msg := <-msgCh:
			c.printMessage(msg)
		default:
		}

		fmt.Print("> ")
		line, err := stdin.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		cmd := parts[0]

		switch cmd {
		case "ping":
			_ = encoder.Encode(GameCmd{Cmd: "ping"})
			var resp GameResp
			_ = decoder.Decode(&resp)
			fmt.Printf("[%s] pong: %s\n", c.playerID, resp.Time)

		case "list":
			_ = encoder.Encode(GameCmd{Cmd: "list"})
			var resp GameResp
			_ = decoder.Decode(&resp)
			fmt.Printf("[%s] Players (%d/%d): %v\n", c.playerID, resp.Count, resp.Max, resp.Players)

		case "chat":
			if len(parts) < 2 {
				fmt.Println("Usage: chat <message>")
				continue
			}
			_ = encoder.Encode(GameCmd{Cmd: "chat", Message: parts[1]})

		case "leave", "quit", "exit":
			_ = encoder.Encode(GameCmd{Cmd: "leave"})
			fmt.Printf("[%s] Left game.\n", c.playerID)
			return nil

		default:
			fmt.Println("Unknown command. Use: ping, list, chat <msg>, leave")
		}
	}

	return nil
}

func (c *Client) printMessage(resp GameResp) {
	switch resp.Cmd {
	case "join":
		fmt.Printf("[SERVER] %s joined\n", resp.Player)
	case "leave":
		fmt.Printf("[SERVER] %s left\n", resp.Player)
	case "chat":
		fmt.Printf("[CHAT] %s: %s\n", resp.From, resp.Message)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// Main
// ═══════════════════════════════════════════════════════════════════════════════

func main() {
	var (
		gateway  = flag.String("gateway", "", "Gateway URL (e.g. http://localhost:8080). If empty, uses -direct")
		direct   = flag.String("direct", "", "Connect directly to game server host:port, skip orchestrator")
		gameID   = flag.Int64("game-id", 1, "Game ID")
		playerID = flag.String("player-id", "player-1", "Player ID")
	)
	flag.Parse()

	if *direct != "" {
		// Direct connect mode — skip orchestrator entirely
		fmt.Printf("[%s] Direct connect to %s\n", *playerID, *direct)
		conn, err := net.Dial("tcp", *direct)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
			os.Exit(1)
		}
		client := NewClient("", *gameID, *playerID)
		if err := client.Play(conn); err != nil {
			fmt.Fprintf(os.Stderr, "Game error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *gateway == "" {
		fmt.Fprintln(os.Stderr, "Either -gateway or -direct must be specified")
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  client -gateway http://localhost:8080 -game-id 1 -player-id alice")
		fmt.Fprintln(os.Stderr, "  client -direct localhost:7777 -player-id bob")
		os.Exit(1)
	}

	client := NewClient(*gateway, *gameID, *playerID)

	conn, server, err := client.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[%s] Connected to server %s:%d (instance %s)\n",
		*playerID, server.Address, server.Port, server.InstanceID)

	if err := client.Play(conn); err != nil {
		fmt.Fprintf(os.Stderr, "Game error: %v\n", err)
	}

	// Try to leave queue if we were in it
	_ = client.QueueLeave()
}
