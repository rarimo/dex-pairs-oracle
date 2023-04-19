-- +migrate Up

drop function if exists trigger_set_updated_at cascade;
CREATE FUNCTION trigger_set_updated_at() RETURNS trigger
    LANGUAGE plpgsql
AS $$ BEGIN NEW.updated_at = NOW() at time zone 'utc'; RETURN NEW; END; $$;

create domain int_256 as numeric not null
    check (value > -(2 ^ 256) and value < 2 ^ 256)
    check (scale(value) = 0);

create table if not exists balances (
  account_address bytea not null,
  token bytea not null,
  chain_id bigint not null,
  amount int_256 not null default 0,
  created_at timestamp with time zone not null default now(),
  updated_at timestamp with time zone not null default now(),
  last_known_block bigint not null default 0,
  primary key (chain_id, token, account_address)
);

create index if not exists balances_token_idx on balances using btree(token);
create index if not exists balances_account_address_idx on balances using btree(account_address);
create index if not exists balances_chain_id_idx on balances using btree(chain_id);

-- +migrate Down

drop index if exists balances_chain_id_idx;
drop index if exists balances_account_address_idx;

drop table if exists balances;

drop domain if exists int_256;

-- insert into balances (account_address, token, chain_id, amount, last_known_block) values
-- ('\x0ed7e52944161450477ee417de9cd3a859b14fd0', '\x000000000ca5171087c18fb271ca844a2370fc0a', 56, 0, 27457598),
-- ('\x0ed7e52944161450477ee417de9cd3a859b14fd0', '\x000000089fb24237da101020ff8e2afd14624687', 56, 0, 27457598),
-- ('\x0ed7e52944161450477ee417de9cd3a859b14fd0', '\x000000ef379ee7f4c051f4b9af901a9219d9ec5c', 56, 0, 27457598),
-- ('\x0ed7e52944161450477ee417de9cd3a859b14fd0', '\x000351d035d8bbf2aa3131ebfecd66fb21836f6c', 56, 0, 27457598),
-- ('\x0ed7e52944161450477ee417de9cd3a859b14fd0', '\x000851476180bfc499ea68450a5327d21c9b050e', 56, 0, 27457598),
-- ('\x0ed7e52944161450477ee417de9cd3a859b14fd0', '\x000c7603cc3de5360c56bb5429f371932675cec7', 56, 0, 27457598),
-- ('\x0ed7e52944161450477ee417de9cd3a859b14fd0', '\x000aa5384e9d90a6cd1d2d4ce3ec605570f7bbbf', 56, 0, 27457598),
-- ('\x0ed7e52944161450477ee417de9cd3a859b14fd0', '\x000c6322df760155bbe4f20f2edd8f4cd35733a6', 56, 0, 27457598),
-- ('\x0ed7e52944161450477ee417de9cd3a859b14fd0', '\x00121c8ee7e214d92b793b6606464d118c6d7074', 56, 0, 27457598),
-- ('\x0ed7e52944161450477ee417de9cd3a859b14fd0', '\x0012365f0a1e5f30a5046c680dcb21d07b15fcf7', 56, 0, 27457598),
-- ('\x0ed7e52944161450477ee417de9cd3a859b14fd0', '\x0019450b0fb021ad2e9f7928101b171272cd537c', 56, 0, 27457598),
-- ('\x0ed7e52944161450477ee417de9cd3a859b14fd0', '\x001d68aff5407f90957c4afa7b7a3cfe4e421bb8', 56, 0, 27457598),
-- ('\x0ed7e52944161450477ee417de9cd3a859b14fd0', '\x0025b42bfc22cbba6c02d23d4ec2abfcf6e014d4', 56, 0, 27457598),
-- ('\x0ed7e52944161450477ee417de9cd3a859b14fd0', '\x002d8563759f5e1eaf8784181f3973288f6856e4', 56, 0, 27457599),
-- ('\x0ed7e52944161450477ee417de9cd3a859b14fd0', '\x002af17a61d3d76ff3629ba2d846e80aae01a1bd', 56, 0, 27457599);