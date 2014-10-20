BEGIN;
CREATE TABLE "artist" (
	    "id" serial NOT NULL PRIMARY KEY,
	    "name" varchar(255) NOT NULL
);
CREATE TABLE "song" (
	    "id" serial NOT NULL PRIMARY KEY,
	    "artist_id" integer NOT NULL REFERENCES "artist" ("id") DEFERRABLE INITIALLY DEFERRED,
	    "name" varchar(255) NOT NULL
);
CREATE TABLE "playlist" (
	    "id" serial NOT NULL PRIMARY KEY,
	    "song_id" integer NOT NULL REFERENCES "song" ("id") DEFERRABLE INITIALLY DEFERRED,
	    "time" timestamp with time zone NOT NULL
);
COMMIT;
