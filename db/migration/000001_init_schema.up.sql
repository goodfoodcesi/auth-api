CREATE EXTENSION IF NOT EXISTS  "uuid-ossp";

CREATE TYPE "USER_ROLE" AS ENUM('ADMIN', 'CLIENT', 'DRIVER');

CREATE TABLE "USER"(
   "user_id" UUID NOT NULL DEFAULT (uuid_generate_v4()),
   "first_name" VARCHAR NOT NULL,
   "last_name" VARCHAR NOT NULL,
   "phone_number" VARCHAR NOT NULL,
   "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   "updated_at" TIMESTAMP NOT NULL,
    "user_role" "USER_ROLE" NOT NULL,
   CONSTRAINT "user_pkey" PRIMARY KEY("user_id")
)