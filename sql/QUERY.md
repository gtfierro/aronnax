# Constructing Queries

## Relational Predicates

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

### `OR`

`OR` is performed by a union between the sets of UUIDs created by evaluating the two predicates involved. We will
be using a right-associative grouping as with the `AND` operator.

As with `AND`, our predicates take the form of:

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

The basic syntax of a union is

```sql
-- optionally can be A.uuid, but every derived
-- table needs an alias, so we do need the "as A"
SELECT uuid FROM
(<select clause 1>) as A
union
(<select clause 2>);
```

We will use a similar grammar rule to the `AND` operator:

```
whereClause :   whereTerm
            |   whereTerm AND whereClause
            |   whereTerm OR whereClause
            ;
```

### `NOT`

Before we discuss how to implement `NOT`, we need to decide on the correct semantics for this. Because all of our
predicates are set operations, we can think of `NOT` as a set inversion. This raises the question: which set are we inverting?
So far, we have only been working with the most *recent* forms of documents, so does a `NOT` invert only the relational
predicate, or the time predicate, or both? It makes the most sense to have `NOT` match only the relational predicate, as follows:

* `WHERE NOT Location/Building = "Soda"`: The non-`NOT` variation matches all streams that have `Soda` as the most recent value
    for their `Location/Building` key. The most obvious choice for how to negate this with `NOT` is to match all streams
    whose most recent value of `Location/Building` is NOT `Soda`, but will not match streams who do not have the key, or
    who have erased their value by writing `null` to it.
* `WHERE NOT Location/Building = "Soda" at "1/1/2014"`: following above, this maintains the same time predicate, so
    this will match all streams whose most immediate value before "1/1/2014" for `Location/Building` is not `Soda`.
* `WHERE NOT Location/Building = "Soda" before "1/1/2014"`:  the choice here is between matching all streams who do not have a single
    value before "1/1/2014" that is `Soda`, or matching all streams who have had any value that is not `Soda` before "1/1/2014". Because
    of the existance of the `for` time operator (which only returns matches if the relational predicate is true for the entire
    specified duration), it makes the most sense to match all streams who have any value before "1/1/2014" for the `Location/Building`
    key that is not `Soda`.

Implementation-wise, I am using a very naive `not in` clause to perform the negation against the set of streams that
match the relational predicate, which seems to work fine, even if it is supposed to be slow. The "correct" way would probably
be using an outer join, which would have to be against the set of all UUIDs that match the same time predicate. This may be what
I use in the future.

## Time-Based Predicates

From [this file](https://github.com/gtfierro/aronnax#queries), we define several time-based predicates that augment the relational predicates
used in the `WHERE` clause:

**All singular timestmpas are exclusive**; that is if there is a single bound, then the expressed time is excluded. Including the time can easily be done with `OR` and `AT`.
**All time ranges are lower-inclusive, upper-exclusive: [time1, time2)**; [for these reasons](http://www.cs.utexas.edu/users/EWD/ewd08xx/EWD831.PDF)

Implemented

| operator | syntax | definition | example |
|----------|--------|------------|---------|
| `FOR`    | `WHERE <relational predicate> FOR <time range>`   | True if the predicate is true *for the entire time range* | `where Room = 410 for (now, now -5min)` |
| `AT`     | `WHERE <relational predicate> AT <timestamp>`     | True if the predicate is true at that point in time | `where Room = 410 at 1447366661s` |
| `HAPPENS BEFORE` | `WHERE <relational predicate> BEFORE <timestamp>` | True if predicate true *at any time* before (not including) the given time. | `where Room = 410 happens before 1447366661s` |
| `HAPPENS AFTER`  | `WHERE <relational predicate> HAPPENS AFTER <timestamp>`  | True if predicate is true after (not including) the current time | `where Room = 410 happens after 1447366661s` |
| `HAPPENS IN`     | `WHERE <relational predicate> HAPPENS IN <time range>`    | True if predicate *becomes* true within the given time range | `where Room = 410 happens in (now, now -5min)` |

Coming Soon:

| operator | syntax | definition | example |
|----------|--------|------------|---------|
| `AFTER`  | `WHERE <relational predicate> AFTER <timestamp>`  | True if predicate is true after (and including) the current time | `where Room = 410 after 1447366661s` |
| `BEFORE`  | `WHERE <relational predicate> AFTER <timestamp>`  | True if predicate is true before (and including) the current time | `where Room = 410 before 1447366661s` |
| `IN`     | `WHERE <relational predicate> IN <time range>`    | True if predicate was true *at any point* within the provided time range | `where Room = 410 in (now, now -5min)` |

These last ones can be approximated with combining them with an `OR` with the same predicate but with a temporal `AT` predicate, e.g.

```sql
-- with syntactic sugar
select * where Location/Room = "410" after "1/1/2014";
-- without
select * where Location/Room = "410" happens after "1/1/2014" or Location/Room="410 at "1/1/2014";
```

---

In this section, we will discuss how to implement those in our SQL expression
compiler. From above, we know that at the core of each of our relational
predicates (such as `Location/City = "Berkeley"`), there is a `SELECT`
statement like

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

The time-portion of the inner `SELECT` statement has been hardcoded to take the
maximum (that is, most-recent) timestamp for all pairs of (`uuid`,`dkey`). The
returned set of document IDs and their keys forms the set of documents to which
we apply our relational predicate. The inner join restricts the evaluation to
just those keys and values that pertain to the most recent form of each
document. For the `in`,`for`,`before`,`at`,`after`,`iafter` time predicates,
we will likely follow very similar constructions that can be dropped in place of
the "most recent" formulation.


### `AT`

This operator is most similar to the default construction, because it only considers a single document.
The other operators will match across many documents.

I believe that implementing this operator is as simple as augmenting the nested `SELECT` containing the
`max(timestamp)` operation with using a `WHERE` clause to restrict the timestamps to be before the given
timestamp.

```sql
select distinct data.uuid
from data
inner join
(
        select distinct uuid, dkey, max(timestamp) as maxtime from data
        where timestamp <= 1234567890
        group by dkey, uuid order by timestamp desc
) sorted
on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
where data.dval is not null
and data.dkey = "Location/City" and data.dval = "Berkeley"
```

### `BEFORE`

This operator should also be simply implemented by removing the `max` operator from the timestamp selector
in the inner nested SELECT clause.

If we do not remove the `group by dkey, uuid`, then we only receive a single
<`key`,`value`,`document`> for our query, which is incorrect.

```sql
select distinct data.uuid
from data
inner join
(
        select distinct uuid, dkey, timestamp as maxtime from data
        where timestamp < 1234567890
        order by timestamp desc
) sorted
on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
where data.dval is not null
and data.dkey = "Location/City" and data.dval = "Berkeley"
```

### `AFTER`

It would first seem that this follows logically from `BEFORE`, using something like:

```sql
select distinct data.uuid
from data
inner join
(
        select distinct uuid, dkey, timestamp as maxtime from data
        where timestamp >= 1234567890
        group by dkey, uuid order by timestamp desc
) sorted
on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
where data.dval is not null
and data.dkey = "Location/City" and data.dval = "Berkeley"
```

but this actually fails to match tags that are still valid: it only matches tags that are applied
after the given timestamp. This may actually be serendipitous, as "only applied after" is a flavor
of query that was not covered by the previous imagining of these operators, and is actually more
helpful, so this is what we use:

```sql
select distinct data.uuid
from data
inner join
(
        select distinct uuid, dkey, timestamp as maxtime from data
        where timestamp >= 1234567890
        order by timestamp desc
) sorted
on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
where data.dval is not null
and data.dkey = "Location/City" and data.dval = "Berkeley"
```

### `IN`

This wants to retrieve all documents between the two times, using lower-bound inclusive,
upper-bound exclusive:

```sql
select distinct data.uuid
from data
inner join
(
        select distinct uuid, dkey, timestamp as maxtime from data
        where timestamp >= 1234567890 and timestamp < 9876543210
        order by timestamp desc
) sorted
on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
where data.dval is not null
and data.dkey = "Location/City" and data.dval = "Berkeley"
```

### `FOR`

This one is tricky, because it involves verifying that the relational predicate is true for
the whole expressed duration. With the exception of this operator, the rest of these
predicate constructions can be handled by rendering directly into the nested
`SELECT` clause.

### Applying `NOT`

Are we going to want to allow users to apply a `not` clause to the time predicates, independent
of the relational predicates? I believe that these offer a sufficient coverage, so for now the
answer is **NO**.
