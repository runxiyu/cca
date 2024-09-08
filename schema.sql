CREATE TABLE courses (
	id INTEGER NOT NULL,
	nmax INTEGER NOT NULL,
	title TEXT,
	ctype TEXT,
	teacher TEXT,
	location TEXT
);
CREATE TABLE users (
	id TEXT PRIMARY KEY NOT NULL,
	name TEXT,
	email TEXT,
	department TEXT
);
CREATE TABLE sessions (
	userid TEXT NOT NULL,
	cookie TEXT PRIMARY KEY NOT NULL,
	expr INTEGER NOT NULL,
	FOREIGN KEY(userid) REFERENCES users(id)
);
