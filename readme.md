# GoXeca

original inspiration:
- [Albrow Jobs](https://github.com/albrow/jobs) - for job processing and execution
- [Olebedev When](https://github.com/olebedev/when) - for time parsing


xeca runs as intended but each processes take about 23-30s to get executed which is slow for simple commands like running a ping or making api calls to check server health,
you can send job requests to the server via the /api/add-job endpoint, and get a list of all the jobs in /api/jobs...

## Todo:
- [ ] understand how go routines work properly
- [ ] fix the jobs execution time
- [ ] handle job chaining and more concurrently
- [ ] build a simple react frontend to interact with the system


## use case
xeca's original use case was for running deployment checks running api health status e.t.c 


## License

MIT
