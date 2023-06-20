package session

import (
	"errors"
	"sync"

	"app/core/domain"
)

type User struct {
	ID domain.SessionID

	PlaylistIDUpvoteChannel            chan *domain.PlaylistIDUpvoteMessage
	PlaylistIDsToListenToUpvoteChanges []domain.PlaylistID
	Lock                               sync.Mutex
}

type Store struct {
	sessions map[domain.SessionID]*User
	lock     sync.Mutex
}

func NewStore() *Store {
	return &Store{
		sessions: make(map[domain.SessionID]*User),
		lock:     sync.Mutex{},
	}
}

func (ss *Store) GetAllUserSessions() map[domain.SessionID]*User {
	ss.lock.Lock()
	defer ss.lock.Unlock()

	return ss.sessions
}

func (ss *Store) GetSession(sessionID domain.SessionID) (*User, error) {
	ss.lock.Lock()
	defer ss.lock.Unlock()

	if session, ok := ss.sessions[sessionID]; ok {
		return session, nil
	}

	return nil, errors.New("session not found")
}

func (ss *Store) AddSession(session *User) {
	ss.lock.Lock()
	defer ss.lock.Unlock()

	ss.sessions[session.ID] = session
}
