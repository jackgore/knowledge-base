CREATE TABLE IF NOT EXISTS organization (
	id SERIAL NOT NULL,
	name VARCHAR(64) NOT NULL,
	created_on DATE NOT NULL,
	is_public BOOLEAN NOT NULL DEFAULT true,
	PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS team (
	id SERIAL NOT NULL,
	org_id INT NOT NULL,
	name VARCHAR(64) NOT NULL,
	created_on DATE NOT NULL,
	is_public BOOLEAN NOT NULL DEFAULT true,
	PRIMARY KEY (id),
	FOREIGN KEY (org_id) REFERENCES organization (id)
);

CREATE TABLE IF NOT EXISTS member_of (
	user_id INT NOT NULL,
	team_id INT NOT NULL,
	PRIMARY KEY (user_id, team_id),
	FOREIGN KEY (user_id) REFERENCES users (id),
	FOREIGN KEY (team_id) REFERENCES team (id)
);

CREATE TABLE IF NOT EXISTS team_of (
	org_id INT NOT NULL,
	team_id INT NOT NULL,
	PRIMARY KEY (org_id, team_id),
	FOREIGN KEY (org_id) REFERENCES organization (id),
	FOREIGN KEY (team_id) REFERENCES team (id)
);

CREATE TABLE IF NOT EXISTS users (
	id SERIAL NOT NULL,
	first_name VARCHAR(64) NOT NULL,
	last_name VARCHAR(64) NOT NULL,
	email VARCHAR(64) NOT NULL,
	username VARCHAR(32) NOT NULL,
	password VARCHAR(64) NOT NULL,
	joined_on DATE NOT NULL,
	PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS post (
	id SERIAL NOT NULL,
	submitted_on DATE NOT NULL,
	title VARCHAR(256) NOT NULL,
	content TEXT NOT NULL,
	author INT NOT NULL,
	views INT NOT NULL DEFAULT 0,
	PRIMARY KEY (id),
	FOREIGN KEY (author) REFERENCES users (id)
);

CREATE TABLE IF NOT EXISTS question (
	id INT NOT NULL,
	PRIMARY KEY (id),
	FOREIGN KEY (id) REFERENCES post (id)
);

CREATE TABLE IF NOT EXISTS followup (
	id SERIAL NOT NULL,
	content TEXT NOT NULL,
	submitted_on DATE NOT NULL,
	author INT NOT NULL,
	PRIMARY KEY (id),
	FOREIGN KEY (author) REFERENCES users (id)
);

CREATE TABLE IF NOT EXISTS answer (
	id INT NOT NULL,
	question INT NOT NULL,
	accepted BOOL NOT NULL DEFAULT false,
	PRIMARY KEY (id),
	FOREIGN KEY (id) REFERENCES followup (id),
	FOREIGN KEY (question) REFERENCES question (id)
);
