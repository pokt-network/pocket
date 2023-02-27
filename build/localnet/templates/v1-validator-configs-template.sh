#!/usr/bin/env bash
# Usage: v1-validator-template.sh number_of_validators

set -Eeuo pipefail
script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd -P)

for ((i = 1; i <= $1; i++)); do
    VALIDATOR_NUMBER=$(printf "%03d" $i) envsubst <"$script_dir/v1-validator-configs.yaml.tpl"
done
