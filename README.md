# faang-stonks

_No affiliation with, but inspired by
[r/Superstonks](https://www.reddit.com/r/Superstonks/)._

## Requirements

* Create a REST API where clients can retrieve stock data for FAANG companies
  * Optionally, clients can request historical data for one or more stocks
* Service should update stock prices every _N_ minutes where _N_ is configurable
* Service and supporting services (e.g., database, web layer, etc.) should be
  packages with Dockerfiles, and a docker-compose.yml file allowing full 
  application instantiation.
* Use git for version control and frequently commit changes.

## Motivation

I want to answer the question, how would I structure a service to meet a set
of requirements with no mention of scale?

The scale ambiguity adds an interesting aspect to this problem. Should the
service be able to run on a $5/mo. VPS? Should it run at Google scale? Maybe
somewhere in between? The code could vary widely depending on the scale.

This project is my attempt to find a happy medium between scalable, agile (the
code), readable (i.e., not clever to the point of illegible code), secure, and 
above all, simple.

## Discussion

The project's requirements don't specify the scale at which this service 
should operate. As such, I took my usual approach: keep things simple yet 
flexible to meet future requirements without the need for a complete code
refactor.

For example, SQLite isn't the best choice for time-series data storage at
scale, but it's good enough for this MVP in the absence of detailed scale
requirements. I've included a stub for InfluxDB to illustrate how I could add
support for it without affecting the rest of the code. I also included an in-
memory implementation.

You'll find a similar pattern for financial data providers. I define an
interface the rest of the code consumes, and then add an implementation of
that interface for my financial data provider (IEX Cloud in this case).

## TODO (aka, it's a work in progress)

1. Give everything a once-over*
2. Goto step 1

* I worked on this in my free time, which meant I didn't have many consecutive
  hours dedicated to writing this code. Therefore, I'm sure some stuff escaped
  my attention with all context switching.