create database weibo;
\c weibo;
create table picinfo(
    pid varchar(20),
    title varchar(255),
    rank int,
    arr int,
    original varchar(255),
    regular varchar(255),
    data varchar(10)
);
