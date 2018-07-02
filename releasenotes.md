- [v0.5.0](#v0.5.0)
- [v0.4.0](#v0.4.0)
- [v0.3.0](#v0.3.0)
- [v0.2.1](#v0.2.1)
- [v0.2.0](#v0.2.0)

## v0.5.0

### Enhancement
* Add pprof #62
* Use RSA standard library #60
* Basic check on the memory #59
* Add a prometheus exporter #57
* Run over old systemd platform #49
* Command line daemon alias #48
* Add notify to the systemd unit #47 

### Bugfixes
* Use a dedicated timeout for each package #64
* Fix the signal reset on SIGs during Stop #58 

### Other
* Update the readme #63
* Introduce ineffassign, golint and misspell #56
* Add SaaS CI examples #50

## v0.4.0

### Enhancement
* New wait command #43

### Bugfixes
* Use a more portable version of listUnits #44

## v0.3.0

### Enhancement
* Introduce the daemon and the reset commands #41
* Allow to delete jobs #38 
* Configurable Kubernetes version setup #37 
* Change the severity of the notify when not running in systemd unit #34 
* Display more state of the runtime #33 

### Bugfixes
* Use an intermediate Certificate Authority for cluster-signing #36 
* Display the adapted config in the logs #35 
* Add aws public and ipv4 detection logic #29 

## v0.2.1

### Bugfixes
* Add aws public and ipv4 detection logic #29

## v0.2.0

### Enhancement
* Job type systemd: display the logs #26

### Bugfixes
* Fallback to AWS hostname metadata #23

### Other
* Disable the reboot strategy #22
* Refactor the Makefile #25
* Provide a release documentation #21
