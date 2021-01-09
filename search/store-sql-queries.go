package search

const wordsTableName = "$prefix$_words_mapping"
const numbersTableName = "$prefix$_numbers_mapping"

const wordsTablesDef = `
create table if not exists $prefix$_words_mapping (
  	token varchar(255) not null,
    field varchar(255) not null,
    objects LONGTEXT not null,
    primary key(token, field)
);
`

const numbersTablesDef = `
create table if not exists $prefix$_numbers_mapping (
  	num bigInt not null,
    field varchar(255) not null,
    objects LONGTEXT not null,
    primary key(num, field)
);
`

const insertWord = `
insert into words_mapping values(?, ?, ?);
`

const appendToWord = `
update words_mapping set field=?, objects=? where token=?;
`

const insertNumber = `
insert into numbers_mapping values(?, ?, ?);
`

const appendToNumber = `
update words_mapping set field=?, objects=? where num=?;
`
