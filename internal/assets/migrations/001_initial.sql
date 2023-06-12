-- +migrate Up

drop function if exists trigger_set_updated_at cascade;
CREATE FUNCTION trigger_set_updated_at() RETURNS trigger
    LANGUAGE plpgsql
AS $$ BEGIN NEW.updated_at = NOW() at time zone 'utc'; RETURN NEW; END; $$;

create domain int_256 as numeric not null
    check (value > -(2 ^ 256) and value < 2 ^ 256)
    check (scale(value) = 0);

create table if not exists balances (
  id bigserial not null primary key,
  account_address bytea not null,
  token bytea not null,
  chain_id bigint not null,
  amount int_256 not null default 0,
  created_at timestamp with time zone not null default now(),
  updated_at timestamp with time zone not null default now(),
  last_known_block bigint not null default 0,
  unique (chain_id, token, account_address)
);

create index if not exists balances_token_idx on balances using btree(token);
create index if not exists balances_account_address_idx on balances using btree(account_address);
create index if not exists balances_chain_id_idx on balances using btree(chain_id);

-- +migrate Down

drop index if exists balances_chain_id_idx;
drop index if exists balances_account_address_idx;

drop table if exists balances;

drop domain if exists int_256;
