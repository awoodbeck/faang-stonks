# faang-stonks

_No affiliation with [r/Superstonks](https://www.reddit.com/r/Superstonks/)._

## Requirements

* Create a REST API where clients can retrieve stock data for FAANG companies
  * Optionally, clients can request historical data for one or more stocks
* Service should update stock prices every _N_ minutes where _N_ is configurable
* Service and supporting services (e.g., database, web layer, etc.) should be
  packages with Dockerfiles, and a docker-compose.yml file allowing full 
  application instantiation.
* Use git for version control and frequently commit changes.

## TODO

* Zap logger
* Lumberjack for log rotation
* SQLite for data storage, though InfluxDB would be a better choice if scale
  was a concern.
* Interfaces: Finance and History
* Goroutine that uses a ticker to pull data from the Provider at regular
  intervals. It can update the DataStore.
  * Attempt to pre-populate historical data upon init.
  * Add option to only update stock prices during market open times.
    * This should be the responsibility of the Provider implementation. The
    goroutine that requests this data shouldn't have to manage that state.
    * The reasoning here is we don't want to incur provider costs when the price
    is unlikely to change.
    * Managing this state in memory should be just fine.
    * Batch calls
* Prometheus for metrics.
* Pprof endpoint on localhost.

## Discussion

The project's requirements didn't specify the scale at which this service 
should operate. As such, I took my usual approach: keep things simple yet 
flexible to meet future requirements without the need for a complete code
refactor.

For example, SQLite isn't the best choice for time-series data storage at
scale, but it's good enough for this MVP in the absence of detailed scale
requirements. I've included a stub for InfluxDB to illustrate how I could add
support for it without affecting the rest of the code.

You'll find a similar pattern for financial data providers. I define an
interface the rest of the code consumes, and then add an implementation of
that interface for my financial data provider (IEX Cloud in this case).