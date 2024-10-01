CREATE TABLE courses (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
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
CREATE TABLE choices (
	id INTEGER GENERATED ALWAYS AS IDENTITY,
	seltime BIGINT NOT NULL, -- microseconds
	userid TEXT NOT NULL,
	courseid INTEGER NOT NULL,
	FOREIGN KEY(userid) REFERENCES users(id),
	FOREIGN KEY(courseid) REFERENCES courses(id)
);
