package utils

import (
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

type atomicSyncMap struct {
	count *atomic.Uint64
	m     *sync.Map
}

func (s *atomicSyncMap) Len() uint {
	return uint(s.count.Load())
}

func (s *atomicSyncMap) Add(key string, value interface{}) {
	s.m.Store(key, value)
	s.count.Add(1)
}

func (s *atomicSyncMap) Remove(key string) {
	s.m.Delete(key)
	s.count.CompareAndSwap(s.count.Load(), s.count.Load()-1)
}

func newAtomicSyncMap() *atomicSyncMap {
	return &atomicSyncMap{
		count: new(atomic.Uint64),
		m:     new(sync.Map),
	}
}

// HOF to iterate over all connections in a map
func forEachConnection(conns *sync.Map) func(func(*websocket.Conn)) {
	return func(callback func(*websocket.Conn)) {
		// k -> connID
		// v -> ws.Conn
		conns.Range(func(_, v any) bool {
			callback(v.(*websocket.Conn))
			return true
		})
	}
}

// HOF to remove a connection from a map
func onRemove(key, connId string, outer, inner *atomicSyncMap) func(func()) bool {
	return func(onInnerEmptycallback func()) bool {
		inner.Remove(connId)
		log.WithField("service", "ConnMap").Info("Removed connection by user ", key)

		if inner.Len() == 0 {
			// unsubscribe from Kafka for this user
			onInnerEmptycallback()
			outer.Remove(key)
			log.WithField("service", "ConnMap").Info("Removed user ", key, " from connMap")
			return true
		}

		return false
	}
}

// ConnMap is a map of userIds to a map of connIds to websocket.Conn
type ConnMap struct {
	outer atomicSyncMap
}

func (cm *ConnMap) Add(
	key string,
	conn *websocket.Conn,
) (func(func(*websocket.Conn)), func(func()) bool, bool) {
	connId := uuid.New().String() // Generate unique ID using Go's uuid package

	// connId -> ws.Conn
	var inner *atomicSyncMap
	isInit := false

	if value, ok := cm.outer.m.Load(key); ok {
		inner = value.(*atomicSyncMap)
		inner.Add(connId, conn)
	} else {
		inner = newAtomicSyncMap()
		inner.Add(connId, conn)
		cm.outer.Add(key, inner)
		isInit = true
	}

	return forEachConnection(inner.m), onRemove(key, connId, &cm.outer, inner), isInit
}

func NewConnMap() *ConnMap {
	return &ConnMap{
		outer: *newAtomicSyncMap(),
	}
}
