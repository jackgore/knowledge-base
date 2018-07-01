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
	"github.com/JonathonGore/knowledge-base/storage"
)

const (
	sessionIDLength = 32
)

// Manager implementation using a go sync map
type SMManager struct {
	cookieName       string   // Name of the cookie we are storing in the users http cookies
	publicCookieName string   // Name of the cookie we are storing in the users http cookies
	sessionMap       sync.Map // Thread safe map for storing our sessions
	maxLifetime      int64    // Expiry time for our sessions
	db               storage.Driver
}

// Creates a new session manager based on the given paramaters
func NewSMManager(cookieName string, maxlifetime int64, db storage.Driver) (*SMManager, error) {
	return &SMManager{cookieName: cookieName, publicCookieName: "kb-public", maxLifetime: maxlifetime, sessionMap: sync.Map{}, db: db}, nil
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

func (m *SMManager) unwrapSession(sid string, obj interface{}) (session.Session, error) {
	var s session.Session

	s, ok := obj.(session.Session)
	if !ok {
		log.Printf("Corrupt value stored in session map for id: %v", sid)
		return s, fmt.Errorf("Corrupt value stored in session map for id: %v", sid)
	}

	return s, nil
}

func (m *SMManager) GetSession(r *http.Request) (session.Session, error) {
	var s session.Session

	if !m.HasSession(r) {
		return s, errors.New("no session cookie in http request")
	}

	cookie, _ := r.Cookie(m.cookieName) // Error can be ignored as this is checked in m.HasSession(r)

	sid, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return s, fmt.Errorf("corrupt value store as session id: %v", err)
	}

	obj, ok := m.sessionMap.Load(sid)
	if ok {
		s, err = m.unwrapSession(sid, obj)
		if err != nil {
			m.sessionMap.Delete(sid)
			m.db.DeleteSession(sid)
			return s, err
		}
		return s, nil
	}

	// If session map is not found in cache we must consult the db
	s, err = m.db.GetSession(sid)
	if err != nil {
		return s, errors.New("unable to get session, likely invalid session id")
	}

	// Now that we have the session from the db store it in our session map
	m.sessionMap.Store(sid, s)

	return s, nil
}

// Determines if there is a session cookie attached to the request
func (m *SMManager) HasSession(r *http.Request) bool {
	cookie, err := r.Cookie(m.cookieName)
	return (err == nil && cookie.Value != "")
}

// SessionStart checks the existence of any sessions related to the current request, and creates a new session if none is found.
func (m *SMManager) SessionStart(w http.ResponseWriter, r *http.Request, username string) (session.Session, error) {
	// TODO: Right now if there is a corrupt value for the cookie it will never be repaired
	if m.HasSession(r) {
		log.Printf("Attempted to start session, but found existing session in request to use")
		return m.GetSession(r)
	}

	log.Printf("No session cookie found for user: %v, creating one now", username)

	sid := generateSessionID()
	s := session.Session{SID: sid, Username: username, ExpiresOn: time.Now().Add(time.Duration(m.maxLifetime) * time.Second)}

	m.sessionMap.Store(sid, s)
	m.db.InsertSession(s)

	// Non-http only cookie
	publicCookie := http.Cookie{Name: m.publicCookieName, Value: url.QueryEscape(username), Path: "/", MaxAge: int(m.maxLifetime)}
	http.SetCookie(w, &publicCookie)

	// HTTP only make it so the cookie is only accessible when sending an http request (so not in javascript)
	cookie := http.Cookie{Name: m.cookieName, Value: url.QueryEscape(sid), Path: "/", HttpOnly: true, MaxAge: int(m.maxLifetime)}
	http.SetCookie(w, &cookie)

	return s, nil
}

// SessionDestroy removes the session stored in the requests cookies.
// Typically called on logout.
func (m *SMManager) SessionDestroy(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" {
		return nil
	}

	// Remove session from cache and database
	m.sessionMap.Delete(cookie.Value)
	m.db.DeleteSession(cookie.Value)

	// Overwrite the current cookie with an expired one
	ec := http.Cookie{Name: m.cookieName, Path: "/", HttpOnly: true, Expires: time.Unix(0, 0), MaxAge: -1}
	epc := http.Cookie{Name: m.publicCookieName, Path: "/", Expires: time.Unix(0, 0), MaxAge: -1}
	http.SetCookie(w, &ec)
	http.SetCookie(w, &epc)

	return nil
}

// GenerateSessionID produces a unique sessionID.
func generateSessionID() string {
	b := make([]byte, sessionIDLength)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
