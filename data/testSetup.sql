INSERT INTO users (first_name, last_name, username, password, email, joined_on)
VALUES ('Test', 'User', 'testuser', 'password', 'test@user.com', 'now');

INSERT INTO session (sid, username, created_on, expires_on)
VALUES ('testsession', 'testuser', 'now', '10/28/2097');

INSERT INTO organization (name, created_on, is_public)
VALUES ('testorg', 'now', false);

INSERT INTO team (org_id, name, created_on, is_public)
VALUES (1, 'default', 'now', false);

INSERT INTO member_of (user_id, org_id)
VALUES (1, 1);
