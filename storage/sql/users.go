package sql

import (
	"log"

	"github.com/JonathonGore/knowledge-base/models/user"
)

/* Inserts the given user into the database.
 * This is an all or nothing insertion.
 *
 * Note: Assumes the password in the user object has already been hashed
 *
 * TODO: This function should return the id of the inserted user
 */
func (d *driver) InsertUser(user user.User) error {
	_, err := d.db.Exec("INSERT INTO users(first_name, last_name, username, password, email, joined_on) VALUES($1, $2, $3, $4, $5, $6)",
		user.FirstName, user.LastName, user.Username, user.Password, user.Email, user.JoinedOn)
	if err != nil {
		log.Printf("Unable to insert user: %v", err)
		return err
	}

	return nil
}

// DeleteUser deletes the user with the given username from the database.
func (d *driver) DeleteUserByUsername(uname string) error {
	u, err := d.GetUserByUsername(uname)
	if err != nil {
		return err
	}

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	// Delete user from all orgs and teams
	_, err = tx.Exec("DELETE FROM member_of WHERE user_id=$1", u.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM member_of_team WHERE user_id=$1", u.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM users WHERE id=$1", u.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

/* Attempts to retrieve the user with the given username from the database.
 * Usually used when attempting to see if a username is attached to a user.
 */
func (d *driver) GetUserByUsername(username string) (user.User, error) {
	user := user.User{}
	err := d.db.QueryRow("SELECT id, first_name, last_name, joined_on, password, email FROM users WHERE username=$1",
		username).Scan(&user.ID, &user.FirstName, &user.LastName, &user.JoinedOn, &user.Password, &user.Email)
	if err != nil {
		log.Printf("User with username %v not found: %v", username, err)
		return user, err
	}

	user.Username = username
	return user, nil
}

/* Gets the user with the given userID from the database.
 */
func (d *driver) GetUser(userID int) (user.User, error) {
	user := user.User{}
	err := d.db.QueryRow("SELECT id, first_name, last_name, joined_on, email FROM users WHERE id=$1",
		userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.JoinedOn, &user.Email)
	if err != nil {
		log.Printf("Unable to retrieve user with id %v: %v", userID, err)
		return user, err
	}

	return user, nil
}
