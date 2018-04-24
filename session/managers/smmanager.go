package managers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/JonathonGore/knowledge-base/session"
)

// Manager implementation using a go sync map
type SMManager struct {
	cookieName  string   // Name of the cookie we are storing in the users http cookies
	sessionMap  sync.Map // Thread safe map for storing our sessions
	maxLifetime int64    // Expiry time for our sessions
}

// Creates a new session manager based on the given paramaters
func NewSMManager(cookieName string, maxlifetime int64) (*SMManager, error) {
	return &SMManager{cookieName: cookieName, maxLifetime: maxlifetime, sessionMap: sync.Map{}}, nil
}

/*
// Session garbage collection -- TODO: If we choose to persist cookies to files this will have to be changed
func (manager *Manager) GC() {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.provider.SessionGC(manager.maxlifetime)
	time.AfterFunc(time.Duration(manager.maxlifetime), func() { manager.GC() })
}
*/

func (m *SMManager) GetSession(r *http.Request) (session.Session, error) {
	var s session.Session

	if !m.HasSession(r) {
		return s, errors.New("could not retrieve requested session")
	}

	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		return s, errors.New("no session cookie in http request")
	}

	sid, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return s, fmt.Errorf("corrupt value store as session id: %v", err)
	}

	isession, ok := m.sessionMap.Load(sid)
	if !ok {
		return s, errors.New("unable to get session, likely invalid session id")
	}

	s, ok = isession.(session.Session)
	if !ok {
		log.Printf("Corrupt value stored in session map for id: %v", sid)
		return s, fmt.Errorf("Corrupt value stored in session map for id: %v", sid)
	}

	return s, nil
}

func (m *SMManager) HasSession(r *http.Request) bool {
	cookie, err := r.Cookie(m.cookieName)
	return (err == nil && cookie.Value != "")
}

// checks the existence of any sessions related to the current request, and creates a new session if none is found.
func (m *SMManager) SessionStart(w http.ResponseWriter, r *http.Request, username string) (session.Session, error) {
	// TODO: Right now if there is a corrupt value for the cookie it will never be repaired
	if m.HasSession(r) {
		log.Printf("Found existing session to use")
		return m.GetSession(r)
	}

	log.Printf("No session cookie found, creating one now")

	sid := m.generateSessionID()
	s := session.Session{SID: sid, Username: username, Expiry: time.Now().Add(time.Duration(m.maxLifetime) * time.Second)}
	m.sessionMap.Store(sid, s)

	// HTTP only make it so the cookie is only accessible when sending an http request (so not in javascript)
	cookie := http.Cookie{Name: m.cookieName, Value: url.QueryEscape(sid), Path: "/", HttpOnly: true, MaxAge: int(m.maxLifetime)}
	http.SetCookie(w, &cookie)

	return s, nil
}

// Destroys the session stored in the requests cookies -- Needs to be called on logout
func (m *SMManager) SessionDestroy(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" {
		log.Printf("Attempted to delete non-existant session")
		return nil
	}

	m.sessionMap.Delete(cookie.Value)
	// Overwrite the current cookie with an expired one
	ec := http.Cookie{Name: m.cookieName, Path: "/", HttpOnly: true, Expires: time.Unix(0, 0), MaxAge: -1}
	http.SetCookie(w, &ec)

	return nil
}

// Produces a unique sessionID
func (m *SMManager) generateSessionID() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
