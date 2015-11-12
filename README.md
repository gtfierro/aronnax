# Aronnax

Aronnax is a document-based database that permits queries over the history of a
document (longitudinal) as well as normal relational queries. I would also like
to support continuous views, in which query result sets are updated in
real-time as the database applies changes to documents.

The goal here is to investigate the construction of such a logging database,
and how the desire to query over history influences the design and
implementation of indexes, views, on-disk storage, query language/API.

## Overview

### Documents and Streams

A `document` is identified by a UUID and consists of a bag of key/value pairs.
Keys are variable-length strings, and values are most likely ints, flots or
strings (maybe lists?).

Aronnax stores `streams`. A `stream` is the history of the key/value pairs in the
bag for a particular UUID. As such, a `document` at time `t` is the keys and
values in the bag at time `t`. Over time, keys may have their values changed,
or may be added (with a value) or removed from a document.

Each edit of a document occurs at some time `t`, so it becomes possible to
query a document as it existed before or after that time. This implements a
timeline of events for a given document. We want to be able to query across all
documents at time `t`, and also query a document over a range of times `t0 < t1`

A naive implementation may implement a document stream by replicating the full
value of a document at every edit timestamp, but it might be better to
implement a stream as a collection of "diffs" created by each document update.
The value of a document at time `t` is then the "sum" or "rollup" of all diffs
up to (and including?) time `t`.

My suspicion is that a log-based approach will reduce conflicts between
concurrent writers by offering a natural serialization (probably order of
delivery) that works with concurrent readers, much in the spirit of data
structures such as the time-ordered multiversion concurrency control mechanisms
described
[here](http://courses.cs.vt.edu/~cs5204/fall07-kafura/Papers/Transactions/ConcurrencyControl.pdf).
However, despite the larger size required for the full-duplication approach, the reads might
be much faster.

### Queries

Want to enable two flavors of queries: vertical and horizontal (probably need better names?).

"Vertical" queries behave like one would expect in a normal document database.
These queries perform relational-type queries over all documents as they exist
at a time `t`, which default to "now". These queries return documents, keys and values.

* This query returns the unique set of UUIDs that matched the given predicate in the last day since the query was written.
    ```sql
    -- find all stream UUIDs that were in 410 Soda in the past day
    select unique uuid where (Metadata/Location/Room = 410 and Metadata/Location/Building = Soda) in (now, now -1d);
    ```

* On a modular sensor platform, the type of temperature sensor was changed from an SHT11 to an SHT13. We want to discover which motes
  had a sensor changed (assuming we know the time of the change)
    ```sql
    select * where Metadata/Sensor/Model = 'SHT11' before 1447364866579373783 and Metadata/Sensor/Model = 'SHT13' after 1447364866579373783;
    ```

Obviously, there are some new semantics that we would like to cover. `where` clauses need an additional qualification that identifies
how the predicates are to be applied over time. These query semantics can be applied to a single clause, or a compound clause (using `and`
or `or`):

* `where key=val in <time range>`, e.g. `where Metadata/Loc/Room = 410 in (now, now -5min)`. True if the predicate was true *at any point*
    within that time range.
* `where key=val for <time range>`, e.g. `where Metadata/Loc/Room = 410 for (now, now -5min)`. True if the predicate is true *for the entire time range*.
* `where key=val before <time>`, e.g. `where Metadata/Loc/Room = 410 before 1447366661s`. True if predicate true *at any time* before the given time.
* `where key=value ibefore <time>`, e.g. `... ibefore 1447366661s`. True if the predicate is true in the most immediate edit before the given time.
* `where key=val after <time>`, e.g. `where Metadata/Loc/Room = 410 after 1447366661s`. True if predicate true *at any time* after the given time.
* `where key=value iafater <time>`, e.g. `... iafter 1447366661s`. True if the predicate is true in the most immediate edit after the given time.

---

"Horizontal" queries operate across the historical values for a key in a document.
These queries return times or ranges of times, augmented with keys or values.



## Data Structures

Data  structure choice is going to be important here. Here are the influencing decisions,
as determined by the queries we want to enable:

* partial matches on strings in both keys and values
    * substring, at the very least (`.*abc`, `abc.*`, `.*abc.*`)
    * maybe case insensitive?
    * maybe full regex? Probably very slow, but let it through anyway. Optimize for the above cases
* grouping keys by the document/stream identifier, sliced by time
    * all keys for document ABC that existed at time `t`
