CREATE TABLE IF NOT EXISTS "jam" (
    "id" uuid PRIMARY KEY DEFAULT uuid_generate_v4 (),
    "name" text NOT NULL CHECK (name <> ''),
    "bpm" int NOT NULL DEFAULT 120 CHECK (bpm > 0),
    "capacity" int NOT NULL DEFAULT 5 CHECK (capacity > 0),
    "owner_id" uuid REFERENCES "user",
    "private" BOOLEAN NOT NULL DEFAULT (false),
    "created_at" timestamptz NOT NULL DEFAULT (now()),
    "updated_at" timestamptz NOT NULL DEFAULT (now()),
    "deleted_at" timestamptz
);

CREATE TABLE IF NOT EXISTS "user" (
    "id" uuid PRIMARY KEY DEFAULT uuid_generate_v4 (),
    "username" text UNIQUE NOT NULL CHECK (username <> ''),
    "email" citext UNIQUE NOT NULL CHECK (email ~ '^[a-zA-Z0-9.!#$%&’*+/=?^_\x60{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*$'),
    "created_at" timestamptz NOT NULL DEFAULT (now()),
    "updated_at" timestamptz NOT NULL DEFAULT (now()),
    "deleted_at" timestamptz
);
