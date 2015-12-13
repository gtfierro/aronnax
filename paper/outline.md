Contextual/Metadata lineage for evolving time series data

Ultimately want to motivate the creation of a database. Paper proves that the characteristics and semantics we need are not offered by current databases. Design and implementation of temporal query language serves to illustrate this.
the difference w/ the aronnax stuff is that the data you are quering has a time-varying context. The timeseries data itself has well-defined temporal characteristicss (discrete samples, mostly). We want to do "as though" queries 

Increasing amoutn of physical information leads to increasing amount of timeseries data
Timeseries data is only so useful as its context: engineering units, location, orientation, calibration
constants, software versions etc -- this is metadata
For making sense of timeseries data, we must make sense of metadata as it changes over time.
There is already a wealth of research on temporal databasees, but these do not server our purposes because:
- overly complicated models -- hard to reason about how to query the data you want
- slow? Not optimized for the most common case: querying most recent version of the database
    - because metadata changes do not happen often (but queries will), we do not want to
      pay for features we are not using: e.g. no rollups over the history!
- the data we are working with is *not* ER data (traditional relational)
- temporal semantics are not apropriate for providing context to a separate database
    - work on phrasing this
- two types: integrated (modify a DBMS to support temporal queries) 

Motivating Applications
- offline analysis:
    - sensor move (floor, room, orientation, timezone)
    - software hanges
    - reporting rates change
    - calibration constants change
    - tagging of transient events: [T1, T2] was a voltage sag
- treat metadata as data
    - all rooms a sensor was in over past month
    - all sensors that were in room XYZ at the time this event happened
- realtime change monitoring
    - control process takes average of all sensors in rooms 1,2,3

Timeseries Metadata:
introduce our structure: streams w/ unique ids, bag of key-value pairs
lends itself to an eav structure, which we can augment with time
With this constraied structure, can we work past the limitations
of past temporal databases?

discuss the data model -- what are the design dimensions?
- how many different times
- where does time come from: client or database?

discuss the implementation

related work:
academic + industry
