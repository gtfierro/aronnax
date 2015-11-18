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
        select distinct uuid, dkey, max(timestamp) as maxtime from data
        group by dkey, uuid order by timestamp desc
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
        select distinct uuid, dkey, max(timestamp) as maxtime from data
        group by dkey, uuid order by timestamp desc
    ) sorted
    on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
    where data.dval is not null
    -- actual WHERE clause
    and (data.dkey = "Location/Room" and data.dval = '410')
)
```

My guess is we can perform a basic SELECT by extending the where clause ("AND dkey in ('key1', 'key2')"), but we can do
this easily enough in the frontend for now.

### Problems with EAV (and more query variations)

It was only at this point that I realized that what I'm actually building is a
historical variation on an Entity-Attribute-Value database model (EAV model
from here on out). EAV has its roots in LISP, evidently, but has since become a
much-maligned antipattern in the relational database community because it becomes used
in cases where a more traditional, strict-schema, relational model fits much
better. EAV models, for databases at least, shine in some niche areas, namely
when column names are not known or change (relatively) frequently.

Saw a post on StackOverflow where a variation of Greenspun's 10th rule was
applied to EAV models: "any sufficiently complex EAV project contains an ad
hoc, informally-specified, bug-ridden, slow implementation of half of a DBMS".

This is the general direction we're headed, but we do not require the full
flexibility of the relational model (at least in practice). This assertion
probably requires some thinking, but we're rolling with it for now.

EAV models offer much flexibility in the kind of information they are able to
represent, but querying them becomes an absolute nightmare due to lots of unions, joins, etc,
which also make the queries much slower than their equivalents on a strictly relational database.

The following query is for  the equivalent of
```sql
select * where Metadata/Exposure = 'South' and Location/Room = '411' and Location/Building = 'Soda'
```

```sql
select second.uuid, second.dkey, second.dval
from (
   select data.uuid, data.dkey, data.dval
   from data
   inner join
   (
        select distinct uuid, dkey, max(timestamp) as maxtime from data
        group by dkey, uuid order by timestamp desc
   ) sorted
   on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
   where data.dval is not null
) as second
right join
(
    select distinct a.uuid from
    (
        select distinct data.uuid
        from data
        inner join
        (
            select distinct uuid, dkey, max(timestamp) as maxtime from data
            group by dkey, uuid order by timestamp desc
        ) sorted
        on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
        where data.dval is not null
            and (data.dkey = "Metadata/Exposure" and data.dval = 'South')
    ) as a
    inner join
    (
        select distinct data.uuid
        from data
        inner join
        (
            select distinct uuid, dkey, max(timestamp) as maxtime from data
            group by dkey, uuid order by timestamp desc
        ) sorted
        on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
        where data.dval is not null
            and (data.dkey = "Location/Room" and data.dval = "411")
    ) b
    on a.uuid = b.uuid
    inner join
    (
        select distinct data.uuid
        from data
        inner join
        (
            select distinct uuid, dkey, max(timestamp) as maxtime from data
            group by dkey, uuid order by timestamp desc
        ) sorted
        on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
        where data.dval is not null
            and (data.dkey = "Location/Building" and data.dval = "Soda")
    ) c
    on b.uuid = c.uuid
) internal
on internal.uuid = second.uuid;
```

