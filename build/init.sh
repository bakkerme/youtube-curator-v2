#!/bin/bash

echo "Initialising AI News Processor"

if [ "$DEBUG_SKIP_CRON" = "true" ]; then
    echo "Debug mode: Skipping cron setup and running main directly"
    /app/main
else
    sed -i "s/\[INSERT\]/$CRON_SCHEDULE/" /etc/cron.d/appcron
    printenv  >> /etc/environment
    
    cat /etc/cron.d/appcron | crontab -
    
    echo "Running at $CRON_SCHEDULE"
    crond -f -L "/dev/stdout"
fi
