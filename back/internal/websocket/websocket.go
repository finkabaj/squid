package websocket

import (
	"net/http"
	"sync"

	"github.com/finkabaj/squid/back/internal/middleware"

	"slices"

	"github.com/finkabaj/squid/back/internal/logger"
	"github.com/gorilla/websocket"
)

type EventType string

const (
	PingEvent                     EventType = "PING"
	PongEvent                     EventType = "PONG"
	UndefinedEvent                EventType = "UNDEFINED"
	ProjectCreatedEvent           EventType = "PROJECT_CREATED"
	ProjectUpdatedEvent           EventType = "PROJECT_UPDATED"
	ProjectDeletedEvent           EventType = "PROJECT_DELETED"
	KanbanColumnCreatedEvent      EventType = "KANBAN_COLUMN_CREATED"
	KanbanColumnUpdatedEvent      EventType = "KANBAN_COLUMN_UPDATED"
	KanbanColumnDeletedEvent      EventType = "KANBAN_COLUMN_DELETED"
	KanbanColumnLabelCreatedEvent EventType = "KANBAN_COLUMN_LABEL_CREATED"
	KanbanColumnLabelUpdatedEvent EventType = "KANBAN_COLUMN_LABEL_UPDATED"
	KanbanColumnLabelDeletedEvent EventType = "KANBAN_COLUMN_LABEL_DELETED"
	KanbanRowCreatedEvent         EventType = "KANBAN_ROW_CREATED"
	KanbanRowUpdatedEvent         EventType = "KANBAN_ROW_UPDATED"
	KanbanRowDeletedEvent         EventType = "KANBAN_ROW_DELETED"
	KanbanRowLabelCreatedEvent    EventType = "KANBAN_ROW_LABEL_CREATED"
	KanbanRowLabelUpdatedEvent    EventType = "KANBAN_ROW_LABEL_UPDATED"
	KanbanRowLabelDeletedEvent    EventType = "KANBAN_ROW_LABEL_DELETED"
	KanbanChecklistCreatedEvent   EventType = "KANBAN_CHECKLIST_CREATED"
	KanbanChecklistDeletedEvent   EventType = "KANBAN_CHECKLIST_DELETED"
	KanbanPointCreatedEvent       EventType = "KANBAN_POINT_CREATED"
	KanbanPointUpdatedEvent       EventType = "KANBAN_POINT_UPDATED"
	KanbanPointDeletedEvent       EventType = "KANBAN_POINT_DELETED"
	KanbanCanCommentEvent         EventType = "KANBAN_CAN_COMMENT"
	KanbanCommendDeletedEvent     EventType = "KANBAN_COMMENT_DELETED"
	KanbanCommentCreatedEvent     EventType = "KANBAN_COMMENT_CREATED"
)

type Event struct {
	Type    EventType `json:"type"`
	Message string    `json:"message"`
	Payload any       `json:"payload"`
}

type Server struct {
	sync.RWMutex
	// userID to ws conn
	Conns    map[string][]*websocket.Conn
	upgrader websocket.Upgrader
}

func NewServer() *Server {
	return &Server{
		Conns: make(map[string][]*websocket.Conn),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// TODO: change when ready for production
				return true
			},
		},
	}
}

func (s *Server) HandleWs(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("failed to upgrade connection")
	}

	user := middleware.UserFromContext(r.Context())

	logger.Logger.Debug().Msgf("new incoming connection from user: %s with client: %s", user.ID, ws.RemoteAddr())

	s.Lock()
	if s.Conns[user.ID] == nil {
		s.Conns[user.ID] = make([]*websocket.Conn, 0)
	}
	s.Conns[user.ID] = append(s.Conns[user.ID], ws)
	s.Unlock()

	s.readLoop(ws, user.ID)
}

func (s *Server) readLoop(ws *websocket.Conn, userID string) {
	defer s.removeConnection(ws, userID)

	for {
		var evt Event
		err := ws.ReadJSON(&evt)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Logger.Error().Err(err).Msg("error reading event")
			}
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

func (s *Server) handlePing(ws *websocket.Conn) {
	evt := Event{
		Type:    PongEvent,
		Message: "pong",
		Payload: nil,
	}

	err := ws.WriteJSON(evt)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("error sending pong")
	}
}

func (s *Server) handleUndefined(ws *websocket.Conn) {
	evt := Event{
		Type:    UndefinedEvent,
		Message: "undefined event",
		Payload: nil,
	}

	err := ws.WriteJSON(evt)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("error sending undefined")
	}
}

func (s *Server) removeConnection(ws *websocket.Conn, userID string) {
	s.Lock()
	defer s.Unlock()

	conns := s.Conns[userID]
	for i, conn := range conns {
		if conn == ws {
			s.Conns[userID] = slices.Delete(conns, i, i+1)
			break
		}
	}

	if len(s.Conns[userID]) == 0 {
		delete(s.Conns, userID)
	}

	ws.Close()
}

func (s *Server) BroadcastToUser(userID string, eventType EventType, msg string, payload any) {
	evt := Event{
		Type:    eventType,
		Message: msg,
		Payload: payload,
	}

	s.RLock()
	connections := s.Conns[userID]
	s.RUnlock()

	for _, conn := range connections {
		if err := conn.WriteJSON(evt); err != nil {
			logger.Logger.Error().Err(err).Msgf("error sending message to user %s", userID)
			s.removeConnection(conn, userID)
		}
	}
}

func (s *Server) BroadcastToProject(projectID string, eventType EventType, msg string, payload any, authorizedUsers []string) {
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
			if err := conn.WriteJSON(evt); err != nil {
				logger.Logger.Error().Err(err).Msgf("error sending message to user %s for project %s", userID, projectID)
				s.removeConnection(conn, userID)
			}
		}
	}
}
