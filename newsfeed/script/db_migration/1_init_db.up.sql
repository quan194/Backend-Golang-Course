create table users
(
    id            bigint NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_name     varchar(20),
    hash_password varchar(60),
    email         varchar(50),
    display_name  nvarchar(20),
    dob           varchar(8),
    removed       boolean
);

create table user_users
(
    id               int NOT NULL AUTO_INCREMENT PRIMARY KEY,
    follower_id      bigint,
    following_id     bigint,
    follow_timestamp int,
    removed          boolean
);
