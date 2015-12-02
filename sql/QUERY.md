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
* `WHERE NOT Location/Building = "Soda" ibefore "1/1/2014"`: following above, this maintains the same time predicate, so
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


| operator | syntax | definition | example |
|----------|--------|------------|---------|
| `IN`     | `WHERE <relational predicate> IN <time range>` | True if predicate was true *at any point* within the provided time range | `where Room = 410 in (now, now -5min)` |
| `FOR`    | `WHERE <relational predicate> FOR <time range>` | True if the predicate is true *for the entire time range* | `where Room = 410 for (now, now -5min)` |
| `BEFORE` | `WHERE <relational predicate> BEFORE <timestamp>` | True if predicate true *at any time* before the given time. | `where Room = 410 before 1447366661s` |
| `IBEFORE`| `WHERE <relational predicate> IBEFORE <timestamp>` | True if the predicate is true in the most immediate edit before the given time. | `where Room = 410 ibefore 1447366661s` |
| `AFTER` | `WHERE <relational predicate> BEFORE <timestamp>` | True if predicate true *at any time* after the given time. | `where Room = 410 after 1447366661s` |
| `IAFTER`| `WHERE <relational predicate> IBEFORE <timestamp>` | True if the predicate is true in the most immediate edit after the given time. | `where Room = 410 iafter 1447366661s` |

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
document. For the `in`,`for`,`before`,`ibefore`,`after`,`iafter` time predicates,
we will likely follow very similar constructions that can be dropped in place of
the "most recent" formulation.


### `IBEFORE`

This operator is most similar to the default construction, because it only
considers a single document. The other operators will match across many
documents.
