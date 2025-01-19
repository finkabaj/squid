package websocket

import (
	"sync"

	"github.com/finkabaj/squid/back/internal/logger"
	"github.com/finkabaj/squid/back/internal/types"
	"golang.org/x/net/websocket"
)

type EventType string

const (
	PingEvent           EventType = "PING"
	PongEvent           EventType = "PONG"
	UndefinedEvent      EventType = "UNDEFINED"
	ProjectCreatedEvent EventType = "PROJECT_CREATED"
)

type Event struct {
	Type    EventType   `json:"type"`
	Message string      `json:"message"`
	Payload interface{} `json:"payload"`
}

type WebsocketServer struct {
	sync.RWMutex
	// userID to ws conn
	Conns map[string][]*websocket.Conn
}

func NewServer() *WebsocketServer {
	return &WebsocketServer{
		Conns: make(map[string][]*websocket.Conn),
	}
}

func (s *WebsocketServer) HandleWs(ws *websocket.Conn) {
	user, ok := ws.Request().Context().Value("user").(*types.User)
	if !ok || user == nil {
		logger.Logger.Error().Msg("unathorized ws connection attemp")
		ws.Close()
		return
	}

	logger.Logger.Debug().Msgf("new incoming connection from user: %s with client: %s", user.ID, ws.RemoteAddr())

	s.Lock()
	if s.Conns[user.ID] == nil {
		s.Conns[user.ID] = make([]*websocket.Conn, 0)
	}
	s.Conns[user.ID] = append(s.Conns[user.ID], ws)
	s.Unlock()

	s.readLoop(ws, user.ID)
}

func (s *WebsocketServer) readLoop(ws *websocket.Conn, userID string) {
	defer s.removeConnection(ws, userID)

	for {
		var evt Event
		err := websocket.JSON.Receive(ws, &evt)
		if err != nil {
			logger.Logger.Error().Err(err).Msg("error reading event")
			return
		}

		switch evt.Type {
		case PingEvent:
			s.handlePing(ws)
		default:
			logger.Logger.Warn().Msgf("unknown event type: %s", evt.Type)
			s.handleUndefined(ws)
		}
	}
}

func (s *WebsocketServer) handlePing(ws *websocket.Conn) {
	evt := Event{
		Type:    PongEvent,
		Message: "pong",
		Payload: nil,
	}

	err := websocket.JSON.Send(ws, evt)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("error sending pong")
	}
}

func (s *WebsocketServer) handleUndefined(ws *websocket.Conn) {
	evt := Event{
		Type:    UndefinedEvent,
		Message: "undefined event",
		Payload: nil,
	}

	err := websocket.JSON.Send(ws, evt)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("error sending undefined")
	}
}

func (s *WebsocketServer) removeConnection(ws *websocket.Conn, userID string) {
	s.Lock()
	defer s.Unlock()

	conns := s.Conns[userID]
	for i, conn := range conns {
		if conn == ws {
			s.Conns[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}

	if len(s.Conns[userID]) == 0 {
		delete(s.Conns, userID)
	}

	ws.Close()
}

func (s *WebsocketServer) BroadcastToUser(userID string, eventType EventType, msg string, payload interface{}) {
	evt := Event{
		Type:    eventType,
		Message: msg,
		Payload: payload,
	}

	s.RLock()
	connections := s.Conns[userID]
	s.RUnlock()

	for _, conn := range connections {
		if err := websocket.JSON.Send(conn, evt); err != nil {
			logger.Logger.Error().Err(err).Msgf("error sending message to user %s", userID)
			s.removeConnection(conn, userID)
		}
	}
}

func (s *WebsocketServer) BroadcastToProject(projectID string, eventType EventType, msg string, payload interface{}, authorizedUsers []string) {
	evt := Event{
		Type:    eventType,
		Message: msg,
		Payload: payload,
	}

	s.RLock()
	defer s.RUnlock()

	for _, userID := range authorizedUsers {
		userConnections, ok := s.Conns[userID]
		if !ok {
			continue
		}
		for _, conn := range userConnections {
			if err := websocket.JSON.Send(conn, evt); err != nil {
				logger.Logger.Error().Err(err).Msgf("error sending message to user %s for project %s", userID, projectID)
				s.removeConnection(conn, userID)
			}
		}
	}
}
