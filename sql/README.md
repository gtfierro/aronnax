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

### Taking Queries Farther

The above query gives us the latest full documents for every UUID, but most where clauses are actually going
to fail us. We can easily filter the results by a UUID-only predicate, but as soon as we incorporate WHERE clauses
that incorporate other keys, the above SQL query only returns the UUID, key and value of the keys mentioned in the
WHERE clause. This is because we are still fundamentally a row-based store.

To actually get the kinds of query semantics we want, we nest the above SQL expression in *another* select clause.
The process is: use the above expression to get a unique set of UUIDs for documents that match the where clause,
and use this as a where clause for another instance of the above expression:

```sql
-- A query for "select * where Location/Room = '410'"
select second.uuid, second.dkey, second.dval
from (
   -- this is the SQL expression from above, unaltered. It gives us the full document
   select data.uuid, data.dkey, data.dval
   from data
   inner join
   (
        select distinct uuid, dkey, max(timestamp) as maxtime from data group by dkey, uuid order by timestamp desc
   ) sorted
   on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
   where data.dval is not null
) as second
where -- filter by documents that match our where clause
uuid in (
    -- this inner query finds the unique set of UUIDs that match our where clause
    select distinct data.uuid
    from data
    inner join
    (
        select distinct uuid, dkey, max(timestamp) as maxtime from data group by dkey, uuid order by timestamp desc
    ) sorted
    on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
    where data.dval is not null
    -- actual WHERE clause
    and (data.dkey = "Location/Room" and data.dval = '410')
)
```

My guess is we can perform a basic SELECT by extending the where clause ("AND dkey in ('key1', 'key2')"), but we can do
this easily enough in the frontend for now.
