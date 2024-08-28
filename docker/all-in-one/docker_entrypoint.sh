#!/bin/bash

# run migrate
/bin/openfga migrate

# run the openfga service in the background
OPENFGA_LOG_FORMAT=json OPENFGA_PLAYGROUND_ENABLED=true /bin/openfga run --experimentals check-query-cache --check-query-cache-enabled &

FGACHECK=1
while [ $FGACHECK -ne 0 ]; do
	grpc_health_probe -addr=:8081
	FGACHECK=$?
done

/bin/redis-server --save 20 1 --loglevel warning --daemonize yes

# run the dbx service in the background if enabled
if [ $CORE_DBX_ENABLED = "true" ]; then
	/bin/dbx serve --debug --pretty &!
fi

# run the core service in the foreground
/bin/core serve --debug --pretty