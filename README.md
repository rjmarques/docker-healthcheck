# Docker Healthcheck App

## Motivation

I wanted to see if it was possible and feasible to user the Docker `HEALTHCHECK` feature to evaluate an app's responsiveness. To this effect I've writen two apps:

- Doctor - Queries the Docker healthchecks (via Docker API) and provides a basic UI to visualise them
- Patient - Has a Burn CPU mode that when enabled uses 100% of the container's available CPU. Additionally, has a healthcheck endpoint (used by Docker) that can be used to measure execution performance

## How it works

### Patient

I wanted to have the patient feature a CPU Burn mode that could be toggled on demand to artificially mimic what would happen, execution wise, when an app's CPU get's too busy. The implemented CPU Burn feature can scale to N available cores (determined by the `cpuset` parameter in `docker-compose.yml`) allowing us to scale the burn to however many logical cores we'd like. For the purposes of my testing 1 core is enough, as I only wanted to check if indeed the healthcheks would become slower as the CPU becomes busier.

Parallel to this the patient app provides a `healthcheck` feature that can be used by Docker to not only check if the app is responding, but also how fast it is responding. I wanted the healthcheck computation to actually do some work, so I made it perform a slow (recursive) Fibonacci computation. Instead of having the user have to set what fib number is slow to compute on their machine, at start-up the patient app calculates increasing Fibonacci numbers until it finds _N_, where _Fib(N)_ takes longer than 400ms to execute on 1 core. _N_ is then used on every healthcheck, which also provides a degree of stability to the measurements.

The patient app provides 3 endpoints via REST API:

- start - starts the CPU burn process
- stop - stops the CPU burn process
- healthcheck - performs the healthcheck (_Fib(N)_ computation)

Note that, the healthcheck endpoint returns a simplistic sequence number that can be used to identify the healthcheck. However, this number gets reset at the start of the patient app, so it cannot be used to uniquely differentiate 1 healthcheck against all past ones, between restarts. This is not an issue for my simplistic app, but potentially would not work in a real production env.

Lastly the patient app does not expose any ports and is solely accessible via the network automatically created by docker-compose.

### Doctor

The doctor app is responsible for collecting metrics from the Docker daemon regarding the healthchecks the latter performed on the patient app. The healthchecks are then aggregated and kept in memory so they can used for analysis later.

To evaluate if the app is running slow or not the doctor collects the following data:

- the fastest healthcheck it found
- the mean duration of last 10 healthchecks

It then considers the patient to be slow if _mean_ is 30% slower than the _fastest_ healthcheck. Naturally, these params were tuned on my machine to give a quick response to toggling burn mode, hence why they're so simplistic. My objective was to prove you can collect and calculate metrics, not to perform statistical analysis on them.

There's also an UI (available on port :8080) that allows users to view the simplistic metrics and well as to toggle the CPU burn mode on the patient app. The UI will give the user very simplistic summary of the state of the patient app.

Note: since I wanted everything to run in Docker I had to mount the docker socket on the doctor container. **This is a bad idea in a production environment!**. But it's good enough for a PoC.

## How to run it

Simply

```bash
./run.sh
```

Then access [localhost:8080](http://localhost:8080) in your browser of choice.

There you'll find the metrics described above and buttons to toggle burn mode on the patient.

For best performance I would leave the app running for 2-3 minutes before toggling burn mode, so the doctor can collect a few healthchecks beforehand.

## Final thoughts (pros & cons)

I only scratched the surface of what you can do with healthchecks. For example you can even write [your own custom ones](https://scoutapm.com/blog/how-to-use-docker-healthcheck), giving you much more freedom in how you evaluate your apps responsiveness and performance.

The advantage of this approach is that the entity collecting the metrics is co-hosted on the target App's machine, bypassing all the issues that come with distributed healthchecking.

However, Docker only saves the last 5 healthchecks so something that fetches healthchecks from Docker daemon via `docker inspect` (like I did), needs to do so faster than the daemon performs them. Perhaps this can be overcome with a custom healthcheck that can save the metrics to a file, or even push them directly to a metrics DB like InfluxDB.

## Disclaimer

This project should _NOT_ be used in a PRODUCTION environment whitout significant improvements. The project was mostly a PoC, developed in a few hours, and so many decisions were made to simplify things, making them more britle.
