alter table products
    add column if not exists accepts_crypto boolean not null default false,
    add column if not exists crypto_pricing_mode text not null default 'disabled',
    add column if not exists crypto_price_usdt numeric(20,8) null;

alter table products
    drop constraint if exists products_crypto_consistency,
    drop constraint if exists products_crypto_price_usdt_check,
    drop constraint if exists products_crypto_pricing_mode_check;

alter table products
    add constraint products_crypto_pricing_mode_check
        check (crypto_pricing_mode in ('disabled', 'fixed_usdt', 'rub_rate')),
    add constraint products_crypto_price_usdt_check
        check (crypto_price_usdt is null or crypto_price_usdt > 0),
    add constraint products_crypto_consistency
        check (
            (accepts_crypto = false and crypto_pricing_mode = 'disabled' and crypto_price_usdt is null)
            or (accepts_crypto = true and crypto_pricing_mode = 'fixed_usdt' and crypto_price_usdt is not null)
            or (accepts_crypto = true and crypto_pricing_mode = 'rub_rate' and crypto_price_usdt is null)
        );

create table if not exists platform_exchange_rates (
    currency_pair text primary key,
    rub_per_usdt numeric(20,8) not null check (rub_per_usdt > 0),
    updated_at timestamptz not null default now()
);
