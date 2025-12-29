#!/bin/bash

# run migrate
/bin/openfga migrate

# run the openfga service in the background
OPENFGA_LOG_FORMAT=json OPENFGA_PLAYGROUND_ENABLED=true OPENFGA_METRICS_ENABLE_RPC_HISTOGRAMS=true /bin/openfga run --experimentals check-query-cache --check-query-cache-enabled &

FGACHECK=1
while [ $FGACHECK -ne 0 ]; do
	grpc_health_probe -addr=:8081
	FGACHECK=$?
done

/bin/redis-server --save 20 1 --loglevel warning --daemonize yes

# run the riverboat service in the background
/bin/riverboat serve --debug --pretty &!

# run the core service in the foreground
/bin/core serve --debug --pretty