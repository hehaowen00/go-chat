create table users (
    user_id text not null,
    username text not null,
    password_hash text not null,
    primary key (user_id),
    unique (username)
);

create table channels (
    channel_id text not null,
    name text not null,
    primary key (channel_id),
    unique (name)
);

create table messages (
    message_id text not null,
    channel_id text not null,
    user_id text not null,
    timestamp integer not null,
    content blob not null,
    primary key (message_id)
);