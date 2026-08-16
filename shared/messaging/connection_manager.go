package messaging

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/trunglq04/goride/shared/contracts"
)

var (
	ErrConnectionNotFound = errors.New("connection not found")
)

// connWrapper is a wrapper around the websocket connection to allow for thread-safe operations
// This is necessary because the websocket connection is not thread-safe
type connWrapper struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type ConnectionManager struct {
	connections map[string]*connWrapper // Local connections storage (userId -> connection)
	mu          sync.RWMutex
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*connWrapper),
	}
}

func (cm *ConnectionManager) Upgrade(c *gin.Context) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (cm *ConnectionManager) Add(id string, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.connections[id] = &connWrapper{
		conn: conn,
		mu:   sync.Mutex{},
	}

	log.Printf("Added connection for user: %s", id)
}

func (cm *ConnectionManager) Remove(id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.connections, id)

	log.Printf("Removed connection for user: %s", id)
}

func (cm *ConnectionManager) Get(id string) (*websocket.Conn, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	wrapper, exists := cm.connections[id]
	if !exists {
		return nil, false
	}

	return wrapper.conn, true
}

func (cm *ConnectionManager) SendMessage(id string, message contracts.WSMessage) error {
	cm.mu.RLock()
	wrapper, exists := cm.connections[id]
	cm.mu.RUnlock()

	if !exists {
		return ErrConnectionNotFound
	}

	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()

	return wrapper.conn.WriteJSON(message)
}
