CREATE TABLE IF NOT EXISTS user (
	id SERIAL NOT NULL,
	first_name VARCHAR(64) NOT NULL,
	last_name VARCHAR(64) NOT NULL,
	username varchar(32) NOT NULL,
	password varchar(64) NOT NULL,
	joined_on DATE NOT NULL,
	PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS post (
	id SERIAL NOT NULL,
	submitted_on DATE NOT NULL,
	title VARCHAR(256) NOT NULL,
	content TEXT NOT NULL,
	author INT NOT NULL,
	PRIMARY KEY (id),
	FOREIGN KEY (author) REFERENCES user (id)
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
	FOREIGN KEY (author) REFERENCES user (id)
);

CREATE TABLE IF NOT EXISTS answer (
	id INT NOT NULL,
	question INT NOT NULL,
	accepted BOOL NOT NULL DEFAULT false,
	PRIMARY KEY (id),
	FOREIGN KEY (id) REFERENCES followup (id),
	FOREIGN KEY (question) REFERENCES question (id)
);
