package sql

import (
	"log"
	"strings"

	"github.com/JonathonGore/knowledge-base/models/organization"
)

// GetOrganization retrieves the org with the given ID from the database.
func (d *driver) GetOrganization(orgID int) (organization.Organization, error) {
	org := organization.Organization{}
	err := d.db.QueryRow("SELECT id, name, created_on, is_public FROM team WHERE id=$1 AND is_deleted=false",
		orgID).Scan(&org.ID, &org.Name, &org.CreatedOn, &org.IsPublic)
	if err != nil {
		return org, err
	}

	return org, nil
}

// DeleteOrganization deletes the org with the given name from the database.
func (d *driver) DeleteOrganization(org string) error {
	// Note: Implementing deletes of an organization without a soft delete is more tricky
	// and could be bad if we instantly wipe out sensitive data.
	// Instead for now we will use an `deleted` column.
	_, err := d.db.Exec("UPDATE organization SET is_deleted=true WHERE name=$1", org)
	if err != nil {
		return err
	}

	return nil
}

// GetOrganizationByName retrieves the requested organization from the database
// by performing a case insensitive search.
func (d *driver) GetOrganizationByName(name string) (organization.Organization, error) {
	org := organization.Organization{}
	err := d.db.QueryRow("SELECT id, name, created_on, is_public, "+
		" (SELECT count(*) FROM member_of WHERE id=org_id)"+
		" FROM organization WHERE upper(name)=$1 AND is_deleted=false",
		strings.ToUpper(name)).Scan(&org.ID, &org.Name, &org.CreatedOn, &org.IsPublic, &org.MemberCount)
	if err != nil {
		log.Printf("Error retriving org by name: %v", err)
		return org, err
	}

	return org, nil
}

// Gets a page of organizations from the database for the given user id
func (d *driver) GetUserOrganizations(uid int) ([]organization.Organization, error) {
	rows, err := d.db.Query("SELECT id, name, created_on, is_public,"+
		" (SELECT count(*) FROM member_of WHERE id=org_id),"+
		" (SELECT count(*) FROM team WHERE team.org_id=organization.id)"+
		" FROM organization JOIN member_of ON (id=member_of.org_id)"+
		" WHERE member_of.user_id=$1 AND is_deleted=false order by name", uid)
	if err != nil {
		log.Printf("Unable to receive organizations from the db: %v", err)
		return nil, err
	}

	orgs := make([]organization.Organization, 0)
	for rows.Next() {
		org := organization.Organization{}
		err := rows.Scan(&org.ID, &org.Name, &org.CreatedOn, &org.IsPublic, &org.MemberCount, &org.TeamCount)
		if err != nil {
			log.Printf("Received error scanning in data from database: %v", err)
			continue
		}
		orgs = append(orgs, org)
	}

	return orgs, err
}

// GetOrganizations retrieves a page of organizations from the database.
// If public is true only public organizations are retrieved.
func (d *driver) GetOrganizations(public bool) ([]organization.Organization, error) {
	rows, err := d.db.Query("SELECT id, name, created_on, is_public,"+
		" (SELECT count(*) FROM member_of WHERE id=org_id),"+
		" (SELECT count(*) FROM team WHERE team.org_id=organization.id)"+
		" FROM organization WHERE is_public=$1 and is_deleted=false"+
		" order by name", public)
	if err != nil {
		log.Printf("Unable to receive organizations from the db: %v", err)
		return nil, err
	}

	orgs := make([]organization.Organization, 0)
	for rows.Next() {
		org := organization.Organization{}
		err := rows.Scan(&org.ID, &org.Name, &org.CreatedOn, &org.IsPublic, &org.MemberCount, &org.TeamCount)
		if err != nil {
			log.Printf("Received error scanning in data from database: %v", err)
			continue
		}
		orgs = append(orgs, org)
	}

	return orgs, err
}

// GetUserOrganizations retrieves a page of organizations from the database for the provided user.
func (d *driver) GetUsernameOrganizations(username string) ([]organization.Organization, error) {
	user, err := d.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	return d.GetUserOrganizations(user.ID)
}

// GetOrganizationMembers retrieves a list of member usernames from the given organization.
func (d *driver) GetOrganizationMembers(org string, admins bool) ([]string, error) {
	adminCheck := ""
	if admins {
		adminCheck = " AND member_of.admin=true"
	}

	rows, err := d.db.Query(
		"SELECT username FROM users, organization, member_of"+
			" WHERE users.id = member_of.user_id AND organization.id = member_of.org_id"+
			" AND organization.name = $1 AND is_deleted=false"+
			adminCheck+
			" ORDER BY username", org)
	if err != nil {
		log.Printf("Unable to receive organization members from the db: %v", err)
		return nil, err
	}

	usernames := make([]string, 0)
	for rows.Next() {
		var username string
		err := rows.Scan(&username)
		if err != nil {
			log.Printf("Received error scanning in data from database: %v", err)
			continue
		}
		usernames = append(usernames, username)
	}

	return usernames, nil
}

// InsertOrgMember insert the given username into the provided org.
func (d *driver) InsertOrgMember(username, org string, isAdmin bool) error {
	u, err := d.GetUserByUsername(username)
	if err != nil {
		return err
	}

	o, err := d.GetOrganizationByName(org)
	if err != nil {
		return err
	}

	tx, err := d.db.Begin()
	if err != nil {
		tx.Rollback()
		log.Printf("Unable to begin transaction: %v", err)
		return err
	}

	_, err = tx.Exec("INSERT INTO member_of(user_id, org_id, admin) VALUES($1, $2, $3)", u.ID, o.ID, isAdmin)
	if err != nil {
		tx.Rollback()
		log.Printf("Unable to insert member into org: %v", err)
		return err
	}

	return tx.Commit()
}

// InsertOrganization creates an organization entry in the database
func (d *driver) InsertOrganization(org organization.Organization) (int, error) {
	err := d.db.QueryRow("INSERT INTO organization(name, created_on, is_public) VALUES($1, $2, $3) returning id;",
		org.Name, org.CreatedOn, org.IsPublic).Scan(&org.ID)
	if err != nil {
		log.Printf("Unable to insert org: %v", err)
		return 0, err
	}

	return org.ID, nil
}
