#!/bin/bash
# Send prompts to stderr to avoid buffering issues when called from other scripts
>&2 echo "Choose queue mode:"
>&2 echo "1) SQS/ElasticMQ (multi-worker, recommended for team/dev)"
>&2 echo "2) Sync (local, single worker, for simple/local dev)"
>&2 read -p "Enter 1 or 2 [1]: " queue_mode
queue_mode=${queue_mode:-1}
if [ "$queue_mode" = "1" ]; then
    echo "sqs"
else
    echo "sync"
fi
