package sql

import (
	"log"

	"github.com/JonathonGore/knowledge-base/models/team"
)

/* Gets the teams for the given org from the database.
 */
func (d *driver) GetTeams(org string) ([]team.Team, error) {
	rows, err := d.db.Query("SELECT team.id, team.org_id, team.name, team.created_on, team.is_public,"+
		" (SELECT count(*) FROM member_of_team WHERE member_of_team.team_id=team.id)"+
		" FROM team JOIN organization on (team.org_id = organization.id)"+
		" WHERE organization.name = $1 AND team.name<>'default'"+
		" order by team.name", org)
	if err != nil {
		log.Printf("Unable to receive teams for org %v from the db: %v", org, err)
		return nil, err
	}

	teams := make([]team.Team, 0)
	for rows.Next() {
		team := team.Team{}
		err := rows.Scan(&team.ID, &team.Organization, &team.Name, &team.CreatedOn, &team.IsPublic, &team.MemberCount)
		if err != nil {
			log.Printf("Received error scanning in data from database: %v", err)
			continue
		}
		teams = append(teams, team)
	}

	return teams, err
}

// GetTeam retrieves the team with the requested id.
func (d *driver) GetTeam(teamID int) (team.Team, error) {
	t := team.Team{}
	err := d.db.QueryRow("SELECT id, org_id, name, created_on, is_public"+
		" (SELECT count(*) FROM member_of_team WHERE member_of_team.team_id=team.id)"+
		" FROM team WHERE id=$1",
		teamID).Scan(&t.ID, &t.Organization, &t.Name, &t.CreatedOn, &t.IsPublic, &t.MemberCount)
	if err != nil {
		log.Printf("Unable to retrieve team with id %v: %v", teamID, err)
		return t, err
	}

	return t, nil
}

// GetTeamByName retrieves the request team name belonging to the given org name.
func (d *driver) GetTeamByName(org, name string) (team.Team, error) {
	t := team.Team{}
	err := d.db.QueryRow("SELECT team.id, team.org_id, team.name, team.created_on, team.is_public,"+
		" (SELECT count(*) FROM member_of_team WHERE member_of_team.team_id=team.id)"+
		" FROM team JOIN organization ON "+
		" (team.org_id = organization.id) WHERE organization.name=$1 and team.name=$2",
		org, name).Scan(&t.ID, &t.Organization, &t.Name, &t.CreatedOn, &t.IsPublic, &t.MemberCount)
	if err != nil {
		log.Printf("Unable to retrieve team with name %v from organization %v: %v", name, org, err)
		return t, err
	}

	return t, nil
}

/* Inserts the given team into the database
 */
func (d *driver) InsertTeam(t team.Team) error {
	_, err := d.db.Exec("INSERT INTO team(org_id, name, created_on, is_public) VALUES($1, $2, $3, $4)",
		t.Organization, t.Name, t.CreatedOn, t.IsPublic)
	if err != nil {
		log.Printf("Unable to insert team: %v", err)
		return err
	}

	return nil
}

// InsertTeamMember insert the given username into the provided org.
func (d *driver) InsertTeamMember(username, org, team string, isAdmin bool) error {
	u, err := d.GetUserByUsername(username)
	if err != nil {
		return err
	}

	t, err := d.GetTeamByName(org, team)
	if err != nil {
		return err
	}

	_, err = d.db.Exec("INSERT INTO member_of_team(user_id, team_id, admin) VALUES($1, $2, $3)", u.ID, t.ID, isAdmin)
	if err != nil {
		return err
	}

	return nil
}
