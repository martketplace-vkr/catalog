create table product_characteristics (
    id bigserial primary key,
    product_id bigint not null references products(id) on delete cascade,
    title text not null,
    sort_order integer not null default 0
);

create index product_characteristics_product_id_idx
    on product_characteristics (product_id, sort_order, id);

create table product_characteristic_attributes (
    id bigserial primary key,
    characteristic_id bigint not null references product_characteristics(id) on delete cascade,
    attribute_name text not null,
    attribute_value text not null,
    sort_order integer not null default 0,
    constraint product_characteristic_attributes_name_not_blank check (btrim(attribute_name) <> ''),
    constraint product_characteristic_attributes_value_not_blank check (btrim(attribute_value) <> ''),
    constraint product_characteristic_attributes_unique_name unique (characteristic_id, attribute_name)
);

create index product_characteristic_attributes_characteristic_id_idx
    on product_characteristic_attributes (characteristic_id, sort_order, id);

