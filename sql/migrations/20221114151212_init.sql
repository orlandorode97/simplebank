-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE Currency AS ENUM (
  'USD',
  'EUR',
  'MXM',
  'CAD',
  'JPY'
);
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE "currencies" (
  "id" bigserial PRIMARY KEY,
  "name" Currency NOT NULL
);

CREATE TABLE "accounts" (
  "id" bigserial PRIMARY KEY,
  "owner" varchar NOT NULL,
  "balance" bigint NOT NULL,
  "currency_id" bigint NOT NULL,
  "createad_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "entries" (
  "id" bigserial PRIMARY KEY,
  "account_id" bigint NOT NULL,
  "amount" bigint NOT NULL,
  "createad_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "transfers" (
  "id" bigserial PRIMARY KEY,
  "from_account_id" bigint NOT NULL,
  "to_account_id" bigint NOT NULL,
  "amount" bigint NOT NULL,
  "createad_at" timestamptz NOT NULL DEFAULT (now())
);



CREATE INDEX ON "accounts" ("owner");

CREATE INDEX ON "entries" ("account_id");

CREATE INDEX ON "transfers" ("from_account_id");

CREATE INDEX ON "transfers" ("to_account_id");

CREATE INDEX ON "transfers" ("from_account_id", "to_account_id");

COMMENT ON COLUMN "entries"."amount" IS 'can be negative or positive';

COMMENT ON COLUMN "transfers"."amount" IS 'must be positive';

ALTER TABLE "accounts" ADD FOREIGN KEY ("currency_id") REFERENCES "currencies" ("id");

ALTER TABLE "entries" ADD FOREIGN KEY ("account_id") REFERENCES "accounts" ("id");

ALTER TABLE "transfers" ADD FOREIGN KEY ("from_account_id") REFERENCES "accounts" ("id");

ALTER TABLE "transfers" ADD FOREIGN KEY ("to_account_id") REFERENCES "accounts" ("id");

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS public.accounts CASCADE;

DROP TABLE IF EXISTS public.currencies CASCADE;

DROP TABLE IF EXISTS public.entries CASCADE;

DROP TABLE IF EXISTS public.transfers CASCADE;
-- +goose StatementEnd
