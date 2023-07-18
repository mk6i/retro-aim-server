package oscar

import "github.com/google/uuid"

type Session struct {
	ID         string
	screenName string
}

type SessionManager struct {
	store map[string]*Session
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		store: make(map[string]*Session),
	}
}

func (s *SessionManager) Retrieve(ID string) (*Session, bool) {
	sess, found := s.store[ID]
	return sess, found
}

func (s *SessionManager) NewSession() (*Session, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	sess := &Session{
		ID: id.String(),
	}
	s.store[sess.ID] = sess
	return sess, nil
}
