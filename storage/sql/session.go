package sql

import (
	"log"

	"github.com/JonathonGore/knowledge-base/session"
)

/* Gets the session with the given sid from the database.
 */
func (d *driver) GetSession(sid string) (session.Session, error) {
	s := session.Session{}
	err := d.db.QueryRow("SELECT sid, username, created_on, expires_on FROM session WHERE sid=$1",
		sid).Scan(&s.SID, &s.Username, &s.CreatedOn, &s.ExpiresOn)
	if err != nil {
		log.Printf("Unable to retrieve session with sid %v: %v", sid, err)
		return s, err
	}

	return s, nil
}

/* Inserts the given session into the database
 */
func (d *driver) InsertSession(s session.Session) error {
	_, err := d.db.Exec("INSERT INTO session(sid, username, created_on, expires_on) VALUES($1, $2, $3, $4)",
		s.SID, s.Username, s.CreatedOn, s.ExpiresOn)
	if err != nil {
		log.Printf("Unable to insert session: %v", err)
		return err
	}

	return nil
}

/* Deletes the session with the sid from the database
 */
func (d *driver) DeleteSession(sid string) error {
	_, err := d.db.Exec("DELETE FROM session WHERE sid=$1", sid)
	if err != nil {
		log.Printf("Unable to delete session: %v", err)
		return err
	}

	return nil
}
