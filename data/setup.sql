CREATE TABLE question (
	id INT NOT NULL,
	submitted_on DATE,
	authored_by INT NOT NULL,
	upvotes INT,
	downvotes INT,
	content TEXT,
	PRIMARY KEY (id)
);

CREATE TABLE author (
	id INT NOT NULL,
	first_name VARCHAR(64),
	last_name VARCHAR(64),
	joined_on DATE,
	PRIMARY KEY (id)
);
