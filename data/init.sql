CREATE TABLE IF NOT EXISTS author (
	id SERIAL NOT NULL,
	first_name VARCHAR(64),
	last_name VARCHAR(64),
	joined_on DATE,
	PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS post (
	id SERIAL NOT NULL,
	submitted_on DATE,
	title VARCHAR(256),
	content TEXT,
	author INT NOT NULL,
	PRIMARY KEY (id),
	FOREIGN KEY (author) REFERENCES author (id)
);

CREATE TABLE IF NOT EXISTS question (
	id INT NOT NULL,
	PRIMARY KEY (id),
	FOREIGN KEY (id) REFERENCES post (id)
);

CREATE TABLE IF NOT EXISTS followup (
	id SERIAL NOT NULL,
	content TEXT,
	submitted_on DATE,
	author INT NOT NULL,
	PRIMARY KEY (id),
	FOREIGN KEY (author) REFERENCES author (id)
);

CREATE TABLE IF NOT EXISTS answer (
	id INT NOT NULL,
	question INT NOT NULL,
	accepted BOOL DEFAULT false,
	PRIMARY KEY (id),
	FOREIGN KEY (id) REFERENCES followup (id),
	FOREIGN KEY (question) REFERENCES question (id)
);
