CREATE TABLE courses (
	id INTEGER PRIMARY KEY,
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
	department TEXT,
	session TEXT,
	expr INTEGER
);
