// Dummy Game Server — заглушка для тестирования оркестрации.
//
// Симулирует реальный игровой сервер:
//   - TCP порт для игроков (JSON-протокол: ping, join, chat, list)
//   - HTTP порт для админки и отчётов (/health, /status, /admin/set-players)
//   - Опциональная интеграция с оркестратором (heartbeat, player count reports)
//
// Использование:
//   go run server.go
//   go run server.go -config config.yaml
//   go run server.go -game-port 7777 -http-port 7778 -max-players 32
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════════
// Config
// ═══════════════════════════════════════════════════════════════════════════════

type Config struct {
	GamePort   int    `json:"game_port"`
	HTTPPort   int    `json:"http_port"`
	MaxPlayers int    `json:"max_players"`
	GameID     int64  `json:"game_id"`
	GameName   string `json:"game_name"`
	LogLevel   string `json:"log_level"`
	LogJSON    bool   `json:"log_json"`
	Orchestrator struct {
		Enabled   bool   `json:"enabled"`
		Gateway   string `json:"gateway_url"`
		GameID    int64  `json:"game_id"`
		Heartbeat int    `json:"heartbeat_sec"`
	} `json:"orchestrator"`
}

func defaultConfig() Config {
	return Config{
		GamePort:   7777,
		HTTPPort:   7778,
		MaxPlayers: 16,
		GameID:     1,
		GameName:   "dummy-game",
		LogLevel:   "info",
		LogJSON:    false,
		Orchestrator: struct {
			Enabled   bool   `json:"enabled"`
			Gateway   string `json:"gateway_url"`
			GameID    int64  `json:"game_id"`
			Heartbeat int    `json:"heartbeat_sec"`
		}{
			Enabled:   false,
			Gateway:   "http://localhost:8080",
			GameID:    1,
			Heartbeat: 5,
		},
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// Game Protocol
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
// Player
// ═══════════════════════════════════════════════════════════════════════════════

type Player struct {
	ID       string
	Name     string
	Conn     net.Conn
	Encoder  *json.Encoder
	JoinedAt time.Time
}

// ═══════════════════════════════════════════════════════════════════════════════
// GameServer
// ═══════════════════════════════════════════════════════════════════════════════

type GameServer struct {
	config      Config
	log         *slog.Logger
	playerCount atomic.Int32
	players     sync.Map // map[string]*Player
	startTime   time.Time
	tcpListener net.Listener
	httpServer  *http.Server
	quit        chan struct{}
	wg          sync.WaitGroup
}

func NewGameServer(cfg Config, log *slog.Logger) *GameServer {
	return &GameServer{
		config:    cfg,
		log:       log,
		startTime: time.Now(),
		quit:      make(chan struct{}),
	}
}

func (s *GameServer) Run() error {
	// TCP game server
	addr := fmt.Sprintf(":%d", s.config.GamePort)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen tcp %s: %w", addr, err)
	}
	s.tcpListener = l

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.acceptLoop()
	}()

	s.log.Info("game server listening", slog.String("addr", addr))

	// HTTP admin server
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /status", s.handleStatus)
	mux.HandleFunc("POST /admin/set-players", s.handleSetPlayers)
	mux.HandleFunc("POST /admin/set-max-players", s.handleSetMaxPlayers)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.HTTPPort),
		Handler: mux,
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Error("http server error", slog.String("error", err.Error()))
		}
	}()

	s.log.Info("admin http listening", slog.Int("port", s.config.HTTPPort))

	// Orchestrator integration
	if s.config.Orchestrator.Enabled {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.orchestratorHeartbeat()
		}()
	}

	return nil
}

func (s *GameServer) Stop() {
	close(s.quit)
	if s.tcpListener != nil {
		_ = s.tcpListener.Close()
	}
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = s.httpServer.Shutdown(ctx)
		cancel()
	}
	s.wg.Wait()
	s.log.Info("server stopped")
}

// ─── TCP Game Server ──────────────────────────────────────────────────────────

func (s *GameServer) acceptLoop() {
	for {
		conn, err := s.tcpListener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				s.log.Warn("accept error", slog.String("error", err.Error()))
				continue
			}
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleConnection(conn)
		}()
	}
}

func (s *GameServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	remote := conn.RemoteAddr().String()
	s.log.Info("player connected", slog.String("remote", remote))

	reader := bufio.NewReader(conn)
	encoder := json.NewEncoder(conn)
	var player *Player

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				s.log.Debug("read error", slog.String("remote", remote), slog.String("error", err.Error()))
			}
			break
		}

		var cmd GameCmd
		if err := json.Unmarshal([]byte(line), &cmd); err != nil {
			_ = encoder.Encode(GameResp{Cmd: "error", OK: false, Error: "invalid json"})
			continue
		}

		resp := s.processCmd(&cmd, conn, encoder, &player)
		if resp != nil {
			_ = encoder.Encode(resp)
		}
		if cmd.Cmd == "leave" {
			break
		}
	}

	if player != nil {
		s.removePlayer(player)
	}

	s.log.Info("player disconnected", slog.String("remote", remote))
}

func (s *GameServer) processCmd(cmd *GameCmd, conn net.Conn, enc *json.Encoder, player **Player) *GameResp {
	switch cmd.Cmd {
	case "ping":
		return &GameResp{Cmd: "pong", OK: true, Time: time.Now().Format(time.RFC3339)}

	case "join":
		name := cmd.Name
		if name == "" {
			name = fmt.Sprintf("player-%d", time.Now().UnixMilli()%10000)
		}
		if int(s.playerCount.Load()) >= s.config.MaxPlayers {
			return &GameResp{Cmd: "join", OK: false, Error: "server full"}
		}
		p := &Player{
			ID:       fmt.Sprintf("%s-%d", name, time.Now().UnixNano()),
			Name:     name,
			Conn:     conn,
			Encoder:  enc,
			JoinedAt: time.Now(),
		}
		s.players.Store(p.ID, p)
		s.playerCount.Add(1)
		*player = p

		s.broadcast(&GameResp{Cmd: "join", OK: true, Player: name, Message: fmt.Sprintf("%s joined", name)}, p.ID)
		s.log.Info("player joined", slog.String("name", name), slog.Int("count", int(s.playerCount.Load())))
		return &GameResp{Cmd: "joined", OK: true, Player: name, Count: int(s.playerCount.Load()), Max: s.config.MaxPlayers}

	case "list":
		var names []string
		s.players.Range(func(_, v any) bool {
			names = append(names, v.(*Player).Name)
			return true
		})
		return &GameResp{Cmd: "players", OK: true, Players: names, Count: int(s.playerCount.Load()), Max: s.config.MaxPlayers}

	case "chat":
		if *player == nil {
			return &GameResp{Cmd: "chat", OK: false, Error: "not joined"}
		}
		if cmd.Message == "" {
			return &GameResp{Cmd: "chat", OK: false, Error: "empty message"}
		}
		s.broadcast(&GameResp{Cmd: "chat", OK: true, From: (*player).Name, Message: cmd.Message}, "")
		return &GameResp{Cmd: "chat", OK: true}

	case "leave":
		if *player != nil {
			s.removePlayer(*player)
			*player = nil
		}
		return &GameResp{Cmd: "left", OK: true}

	case "status":
		return &GameResp{Cmd: "status", OK: true, Count: int(s.playerCount.Load()), Max: s.config.MaxPlayers}

	default:
		return &GameResp{Cmd: "error", OK: false, Error: "unknown command: " + cmd.Cmd}
	}
}

func (s *GameServer) removePlayer(p *Player) {
	s.players.Delete(p.ID)
	s.playerCount.Add(-1)
	s.broadcast(&GameResp{Cmd: "leave", OK: true, Player: p.Name, Message: fmt.Sprintf("%s left", p.Name)}, p.ID)
	s.log.Info("player left", slog.String("name", p.Name), slog.Int("count", int(s.playerCount.Load())))
}

func (s *GameServer) broadcast(msg *GameResp, excludeID string) {
	s.players.Range(func(_, v any) bool {
		p := v.(*Player)
		if p.ID != excludeID {
			_ = p.Encoder.Encode(msg)
		}
		return true
	})
}

// ─── HTTP Handlers ────────────────────────────────────────────────────────────

func (s *GameServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"game":   s.config.GameName,
	})
}

func (s *GameServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	var names []string
	s.players.Range(func(_, v any) bool {
		names = append(names, v.(*Player).Name)
		return true
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"status":         "ok",
		"game_id":        s.config.GameID,
		"game_name":      s.config.GameName,
		"player_count":   s.playerCount.Load(),
		"max_players":    s.config.MaxPlayers,
		"uptime_seconds": int(time.Since(s.startTime).Seconds()),
		"game_port":      s.config.GamePort,
		"http_port":      s.config.HTTPPort,
		"players":        names,
	})
}

func (s *GameServer) handleSetPlayers(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	old := s.playerCount.Load()
	s.playerCount.Store(int32(req.Count))
		s.log.Info("admin set player count", slog.Int("old", int(old)), slog.Int("new", req.Count))
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "player_count": req.Count})
}

func (s *GameServer) handleSetMaxPlayers(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Max int `json:"max"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	if req.Max < 1 {
		http.Error(w, `{"error":"max must be >= 1"}`, http.StatusBadRequest)
		return
	}
	old := s.config.MaxPlayers
	s.config.MaxPlayers = req.Max
	s.log.Info("admin set max players", slog.Int("old", old), slog.Int("new", req.Max))
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "max_players": req.Max})
}

// ─── Orchestrator Integration ─────────────────────────────────────────────────

func (s *GameServer) orchestratorHeartbeat() {
	ticker := time.NewTicker(time.Duration(s.config.Orchestrator.Heartbeat) * time.Second)
	defer ticker.Stop()

	gameID := s.config.Orchestrator.GameID
	if gameID == 0 {
		gameID = s.config.GameID
	}

	for {
		select {
		case <-s.quit:
			return
		case <-ticker.C:
			s.reportToOrchestrator(gameID)
		}
	}
}

func (s *GameServer) reportToOrchestrator(gameID int64) {
	gw := s.config.Orchestrator.Gateway
	if gw == "" {
		return
	}

	// 1. Health check gateway
	resp, err := http.Get(gw + "/api/v1/health")
	if err != nil {
		s.log.Warn("orchestrator health check failed", slog.String("error", err.Error()))
		return
	}
	_ = resp.Body.Close()

	// 2. Log report (в реальном сценарии здесь был бы POST к game-server-node или напрямую к оркестратору)
	s.log.Info("report to orchestrator",
		slog.Int64("game_id", gameID),
		slog.Int("player_count", int(s.playerCount.Load())),
		slog.Int("max_players", s.config.MaxPlayers),
		slog.String("gateway", gw),
	)
}

// ═══════════════════════════════════════════════════════════════════════════════
// Helpers
// ═══════════════════════════════════════════════════════════════════════════════

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func setupLogger(level string, jsonOut bool) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: lvl}
	if jsonOut {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}

// ═══════════════════════════════════════════════════════════════════════════════
// Main
// ═══════════════════════════════════════════════════════════════════════════════

func main() {
	var (
		configPath = flag.String("config", "", "path to config file (optional)")
		gamePort   = flag.Int("game-port", 0, "TCP game port (overrides config)")
		httpPort   = flag.Int("http-port", 0, "HTTP admin port (overrides config)")
		maxPlayers = flag.Int("max-players", 0, "max players (overrides config)")
		orch       = flag.Bool("orch", false, "enable orchestrator integration")
		gateway    = flag.String("gateway", "", "orchestrator gateway URL")
		gameID     = flag.Int64("game-id", 0, "game ID")
	)
	flag.Parse()

	cfg := defaultConfig()

	if *configPath != "" {
		data, err := os.ReadFile(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read config: %v\n", err)
			os.Exit(1)
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			// try yaml-like simple parsing: just ignore, use defaults + flags
			_ = err
		}
	}

	// Env vars override
	if v := os.Getenv("DUMMY_GAME_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.GamePort)
	}
	if v := os.Getenv("DUMMY_HTTP_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.HTTPPort)
	}
	if v := os.Getenv("DUMMY_MAX_PLAYERS"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.MaxPlayers)
	}
	if v := os.Getenv("DUMMY_GAME_ID"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.GameID)
	}
	if v := os.Getenv("DUMMY_ORCH_ENABLED"); v == "true" || v == "1" {
		cfg.Orchestrator.Enabled = true
	}
	if v := os.Getenv("DUMMY_ORCH_GATEWAY"); v != "" {
		cfg.Orchestrator.Gateway = v
	}

	// Flags override
	if *gamePort != 0 {
		cfg.GamePort = *gamePort
	}
	if *httpPort != 0 {
		cfg.HTTPPort = *httpPort
	}
	if *maxPlayers != 0 {
		cfg.MaxPlayers = *maxPlayers
	}
	if *orch {
		cfg.Orchestrator.Enabled = true
	}
	if *gateway != "" {
		cfg.Orchestrator.Gateway = *gateway
	}
	if *gameID != 0 {
		cfg.GameID = *gameID
		cfg.Orchestrator.GameID = *gameID
	}

	log := setupLogger(cfg.LogLevel, cfg.LogJSON)

	log.Info("starting dummy game server",
		slog.Int("game_port", cfg.GamePort),
		slog.Int("http_port", cfg.HTTPPort),
		slog.Int("max_players", cfg.MaxPlayers),
		slog.Int64("game_id", cfg.GameID),
		slog.Bool("orchestrator", cfg.Orchestrator.Enabled),
	)

	server := NewGameServer(cfg, log)
	if err := server.Run(); err != nil {
		log.Error("failed to start", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Wait for signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Info("shutting down...")
	server.Stop()
}
