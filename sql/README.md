SQL implementation to explore the semantics

## Test User
Unit tests should probably be run on a separate test database.

```sql
CREATE DATABASE aronnaxtest;
CREATE USER 'aronnaxtest'@'%' IDENTIFIED BY 'aronnaxpass';
GRANT ALL ON aronnaxtest.* TO 'aronnaxtest'@'%';
FLUSH PRIVILEGES;
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
    uuid CHAR(16) NOT NULL,
    dkey VARCHAR(20) NOT NULL,
    dval VARCHAR(20) NULL,
    timestamp TIMESTAMP NOT NULL
);
```
