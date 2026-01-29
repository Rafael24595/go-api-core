package session

type RepositorySession interface {
	FindAll() []Session
	Find(user string) (*Session, bool)
	Insert(session *Session) *Session
	Delete(session *Session) *Session
}
