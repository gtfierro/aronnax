# JFSeb

JFSeb is a document-based database that permits queries over the history of a
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

JFSeb stores `streams`. A `stream` is the history of the key/value pairs in the
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
