# faang-stonks

_No affiliation with r/Superstonks_

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
* InfluxDB for time-series data storage.
* Interfaces: Provider and DataStore
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
* Embed static assets using go:embed.
* Add versioning using go:generate.
* As time allows, use mini.css for front end.
  * If user requests historical prices, chart them.