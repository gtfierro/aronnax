## Constructing Queries

`WHERE` clauses are constructed by evaluating each individual "term" (a single
predicate, e.g. `Location/City = "Berkeley"`) to create sets of document
identifiers (UUIDs), and then taking union/intersections of those to create the
`or` and `and` elements of the `WHERE` clause. We then perform a `RIGHT JOIN`
against this to get the most recent key/value pairs for the documents that match
our `WHERE` clause. This takes the form of:

```sql
select second.uuid, second.dkey, second.dval
from (
   select data.uuid, data.dkey, data.dval
   from data
   inner join
   (
        select distinct uuid, dkey, max(timestamp) as maxtime
        from data
        group by dkey, uuid order by timestamp desc
   ) sorted
   on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
   where data.dval is not null
) as second
right join
(
    -- where clause stuff here
) internal
on internal.uuid = second.uuid;
```

Currently, these queries are focusing on the "most recent" form of these
documents. Once the basic query logic is in place, we should be able to extend
these principles to apply to time-based predicates.

### `AND`

Let's work through the example of `select * where Location/Building = "Soda" and Location/City = "Berkeley";`

AND is performed using an inner join between the sets of UUIDs created by evaluating the predicates `Location/Building="Soda"`
and `Location/City="Berkeley"`. Those each look like

```sql
select distinct data.uuid
from data
inner join
(
        select distinct uuid, dkey, max(timestamp) as maxtime from data
        group by dkey, uuid order by timestamp desc
) sorted
on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
where data.dval is not null
and data.dkey = "Location/City" and data.dval = "Berkeley"
```

which we will abbreviate as `<TERM>`. `AND` is an intersection, so we use a SQL inner join to combine
our two `SELECT` statements. The basic syntax of that is:

```sql
-- this can actually be either A.uuid or B.uuid
-- because we'll get equivalence from the inner join
SELECT A.uuid FROM
(<select clause 1>) as A
inner join
(<select clause 2>) as B
on
A.uuid = B.uuid;
```

For multiple `AND` statements, this is easily chainable:

```sql
SELECT A.uuid FROM
(<select clause 1>) as A
inner join
(<select clause 2>) as B
on A.uuid = B.uuid
inner join
(<select clause 3>) as C
on B.uuid = C.uuid
inner join
(<select clause 4>) as D
on C.uuid = D.uuid;

-- or, using our <TERM>,
SELECT A.uuid FROM
(<TERM1>) as A
inner join
(<TERM2>) as B
on A.uuid = B.uuid
inner join
(<TERM3>) as C
on B.uuid = C.uuid
inner join
(<TERM4>) as D
on C.uuid = D.uuid;
```

This implies that for a grammar rule such as

```
whereClause :   whereTerm
            |   whereTerm AND whereClause
            ;
-- gives us (A and (B and (C and D)))
```

we want to generate something like

```
select A.uuid
( $1 ) as A
inner join
( $2 ) as B
on A.uuid = B.uuid
```

We have a quick function that gives us a new letter every time we call it, so
we can generate new names for these ephemeral tables (created by the `SELECT`
clauses). We only want to generate a new letter for `$1`, because this is the first
time we've seen it.


```
whereClause: whereTerm
            {
                $1.Letter = nextletter()
                // if whereTerm is just a predicate:
                if ($1.IsPredicate) {
                  $$ = whereClause{
                  Select: `
                    select distinct data.uuid
                    from data
                    inner join
                    (
                        select distinct uuid, dkey, max(timestamp) as maxtime from data
                        group by dkey, uuid order by timestamp desc
                    ) sorted
                    on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
                    where data.dval is not null
                    `
                  }
                } else { // it came from ( whereClause ), e.g. we have a full SQL statement so we pass it through
                    $$ = whereClause{Select: $1.Select}
                }
            }
            | whereTerm AND whereClause
            {
                $$ = whereClause{
                    Select: `select $1.uuid
                    from
                    ($1) as $1.Letter
                    inner join
                    ($2) as $2.Letter
                    on $1.uuid = $2.uuid`
                    }
            }
            ;
```

This should recursively generate our AND clauses.
