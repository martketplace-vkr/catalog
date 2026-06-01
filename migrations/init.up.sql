create table categories (
    id bigserial primary key,
    name text not null,
    parent_id bigint references categories(id),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    constraint categories_name_not_blank check (btrim(name) <> '')
);

create unique index categories_parent_name_uidx
    on categories (coalesce(parent_id, 0), lower(name));

create index categories_parent_id_idx
    on categories (parent_id);

create table products (
    id bigserial primary key,
    vendor_id bigint not null,
    category_id bigint not null references categories(id),
    name text not null,
    description text not null default '',
    price numeric(10,2) not null check (price >= 0),
    accepts_crypto boolean not null default false,
    crypto_pricing_mode text not null default 'disabled'
        check (crypto_pricing_mode in ('disabled', 'fixed_usdt', 'rub_rate')),
    crypto_price_usdt numeric(20,8) null check (crypto_price_usdt is null or crypto_price_usdt > 0),
    stock_count integer not null default 0 check (stock_count >= 0),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    constraint products_name_not_blank check (btrim(name) <> ''),
    constraint products_crypto_consistency check (
        (accepts_crypto = false and crypto_pricing_mode = 'disabled' and crypto_price_usdt is null)
        or (accepts_crypto = true and crypto_pricing_mode = 'fixed_usdt' and crypto_price_usdt is not null)
        or (accepts_crypto = true and crypto_pricing_mode = 'rub_rate' and crypto_price_usdt is null)
    )
);

create index products_vendor_id_idx
    on products (vendor_id);

create index products_category_id_idx
    on products (category_id);

create table platform_exchange_rates (
    currency_pair text primary key,
    rub_per_usdt numeric(20,8) not null check (rub_per_usdt > 0),
    updated_at timestamptz not null default now()
);

create table product_attributes (
    id bigserial primary key,
    product_id bigint not null references products(id) on delete cascade,
    attribute_name text not null,
    attribute_value text not null,
    constraint product_attributes_name_not_blank check (btrim(attribute_name) <> ''),
    constraint product_attributes_value_not_blank check (btrim(attribute_value) <> ''),
    constraint product_attributes_product_name_unique unique (product_id, attribute_name)
);

create index product_attributes_product_id_idx
    on product_attributes (product_id);

create table product_images (
    id bigserial primary key,
    product_id bigint not null references products(id) on delete cascade,
    url text not null,
    is_main boolean not null default false,
    constraint product_images_url_not_blank check (btrim(url) <> '')
);

create index product_images_product_id_idx
    on product_images (product_id);

create unique index product_images_one_main_per_product_uidx
    on product_images (product_id)
    where is_main;
