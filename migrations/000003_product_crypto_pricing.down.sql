drop table if exists platform_exchange_rates;

alter table products
    drop constraint if exists products_crypto_consistency,
    drop constraint if exists products_crypto_price_usdt_check,
    drop constraint if exists products_crypto_pricing_mode_check,
    drop column if exists crypto_price_usdt,
    drop column if exists crypto_pricing_mode,
    drop column if exists accepts_crypto;
