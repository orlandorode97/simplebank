-- +goose Up
-- +goose StatementBegin
CREATE TABLE "users" (
  "username" varchar PRIMARY KEY NOT NULL,
  "hashed_password" varchar NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "createad_at" timestamptz NOT NULL DEFAULT (now())
);

ALTER TABLE "accounts" ADD FOREIGN KEY ("owner") REFERENCES "users" ("username");

ALTER TABLE "accounts" ADD CONSTRAINT "owner_currency_key" UNIQUE("owner", "currency_id");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS public.accounts DROP CONSTRAINT IF EXISTS "owner_currency_key";
ALTER TABLE IF EXISTS public.accounts DROP CONSTRAINT IF EXISTS "accounts_owner_fkey";
DROP TABLE IF EXISTS public.users CASCADE;
-- +goose StatementEnd
