#!/bin/bash
# Send prompts to stderr to avoid buffering issues when called from other scripts
>&2 echo "Choose logging channel:"
>&2 echo "1) Sentry (recommended for team/dev)"
>&2 echo "2) Local (logs to file only)"
>&2 read -p "Enter 1 or 2 [1]: " log_mode
log_mode=${log_mode:-1}
if [ "$log_mode" = "1" ]; then
    echo "sentry"
else
    echo "local"
fi
