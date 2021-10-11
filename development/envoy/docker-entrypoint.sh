#!/usr/bin/env sh
set -e

loglevel="${loglevel:-}"
USERID=$(id -u)

# If present /etc/envoy/envoy.yaml.tmpl - generate config /etc/envoy/envoy.yaml from it.
# Fail if any vars are not resolved.
TEMPLATE_FILE="/etc/envoy/envoy.yaml.tmpl"
CONFIG_FILE="/etc/envoy/envoy.yaml"
if [ -f "${TEMPLATE_FILE}" ]; then
    echo "Found $TEMPLATE_FILE, generating $CONFIG_FILE from it"
    gomplate --file "${TEMPLATE_FILE}" --out "${CONFIG_FILE}" || {
        echo "ERROR running gomplate, failing"
        exit 1
    }
    echo "Finished generating $CONFIG_FILE"
fi

# if the first argument look like a parameter (i.e. start with '-'), run Envoy
if [ "${1#-}" != "$1" ]; then
    set -- envoy "$@"
fi

if [ "$1" = 'envoy' ]; then
    # set the log level if the $loglevel variable is set
    if [ -n "$loglevel" ]; then
        set -- "$@" --log-level "$loglevel"
    fi
fi

if [ "$ENVOY_UID" != "0" ] && [ "$USERID" = 0 ]; then
    if [ -n "$ENVOY_UID" ]; then
        usermod -u "$ENVOY_UID" envoy
    fi
    if [ -n "$ENVOY_GID" ]; then
        groupmod -g "$ENVOY_GID" envoy
    fi
    # Ensure the envoy user is able to write to container logs
    chown envoy:envoy /dev/stdout /dev/stderr
    exec su-exec envoy "${@}"
else
    exec "${@}"
fi
