SQL implementation to explore the semantics

## Test User
Unit tests should probably be run on a separate test database.

```sql
CREATE DATABASE aronnaxtest;
CREATE USER 'aronnaxtest'@'%' IDENTIFIED BY 'aronnaxpass';
GRANT ALL ON aronnaxtest.* TO 'aronnaxtest'@'%';
FLUSH PRIVILEGES;
```

## Test the REPL

```bash
go build
ARONNAXUSER=aronnaxtest ARONNAXPASS=aronnaxpass ARONNAXDB=aronnaxtest rlwrap ./sql
```

## Schema

We are using a single SQL table for now with the following columns:

* UUID: 128-bit unique value (`CHAR(16)`) nonnull
* key: varchar(20) nonnull
* value: varchar(20) null
* timestamp: datetime? unsigned int?

```sql
CREATE TABLE data
(
    uuid CHAR(37) NOT NULL,
    dkey VARCHAR(128) NOT NULL,
    dval VARCHAR(128) NULL,
    timestamp TIMESTAMP NOT NULL
);
```

## Queries

To get the timestamp of the most recent change for each key, use

```sql
select distinct uuid, dkey, max(timestamp) from data group by dkey order by timestamp desc;
```

Then to get the actual values, we want

```sql
select data.uuid, data.dkey, data.dval
from data
inner join
(
    select distinct uuid, dkey, max(timestamp) as maxtime from data group by dkey, uuid order by timestamp desc
) sorted
on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
where data.dval is not null;
```

We should be able to augment these with appropriate where clauses for filtering on documents
