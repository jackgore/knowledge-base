INSERT INTO users (first_name, last_name, username, password, email, joined_on)
VALUES ('Test', 'User', 'testuser', 'password', 'test@user.com', 'now');

INSERT INTO session (sid, username, created_on, expires_on)
VALUES ('testsession', 'testuser', 'now', '10/28/2097');
