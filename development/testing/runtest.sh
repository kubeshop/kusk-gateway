#!/usr/bin/env bash

set -eou pipefail

APP="$(basename "$0" .sh)"

NAMESPACE=testing
TESTING_EXTERNAL_IP=""

# Detect where we run script from
if [[ -d "postman" ]]; then
    TESTING_DIRECTORY="${PWD}"
else
    TESTING_DIRECTORY="${PWD}/development/testing"
fi

########################## Subroutines ############################################################
function error() {
    TIMESTAMP="$(date "+%F %T")"
    echo "${TIMESTAMP} ERROR [${APP}] $*" >&2
}

function warn() {
    TIMESTAMP="$(date "+%F %T")"
    echo "${TIMESTAMP} WARN [${APP}] $*" >&2
}

function info() {
    TIMESTAMP="$(date "+%F %T")"
    echo "${TIMESTAMP} INFO [${APP}] $*"
}

function deploy() {
    echo "INFO: Deploying the required manifests"
    kubectl apply -f "${TESTING_DIRECTORY}/manifests"
    wait_for_deployed
    wait_for_eip
}

function wait_for_deployed() {
    echo "INFO: Waiting for deployed resources"
    for I in todo-backend todo-frontend kgw-envoy-testing; do
        kubectl wait --for=condition=available --timeout=60s deployment/"$I" -n "${NAMESPACE}"
    done
}

function get_external_ip() {
    kubectl get svc -l "app.kubernetes.io/component=envoy-svc" -n $NAMESPACE -o=jsonpath='{.items[0].status.loadBalancer.ingress[0].ip}'
}

function wait_for_eip() {
    CHECK_TIMES=10
    while true; do
        TESTING_EXTERNAL_IP=$(get_external_ip)
        if [[ -z "${TESTING_EXTERNAL_IP}" ]]; then
            warn "Haven't found Envoy service External IP yet"
        else
            break
        fi
        CHECK_TIMES=-1
        if [[ ${CHECK_TIMES} == 0 ]]; then
            error "Couldn't find Envoy service External IP, exiting"
            exit 1
        fi
        sleep 1
    done
}

function test_api() {
    info "Testing API"
    TESTING_EXTERNAL_IP=$(get_external_ip)
    docker run --network host --rm -v "${TESTING_DIRECTORY}/postman":/tests -t postman/newman run /tests/api.postman_collection.json --env-var EXTERNAL_IP="${TESTING_EXTERNAL_IP}"
    info "SUCCESS Testing API"
}

function test_staticroute() {
    info "Testing StaticRoutes"
    TESTING_EXTERNAL_IP=$(get_external_ip)
    docker run --network host --rm -v "${TESTING_DIRECTORY}/postman":/tests -t postman/newman run /tests/staticroute.postman_collection.json --env-var EXTERNAL_IP="${TESTING_EXTERNAL_IP}"
    info "SUCCESS Testing StaticRoutes"
}

function get_manager_logs() {
    info "Retrieving manager logs for you with filtered out INFO to check for any errors"
    kubectl logs -l app.kubernetes.io/component=kusk-gateway-manager -n kusk-system -c manager --tail=100 | sed '/INFO/d'
}

function get_envoy_logs() {
    info "Retrieving Envoy logs for you with filtered out [info] to check for any errors"
    kubectl logs -l app.kubernetes.io/component=envoy -n ${NAMESPACE} -c envoy --tail=100 | sed '/\[info\]/d'
}

function delete() {
    info "Deleting created resources"
    kubectl delete -f "${TESTING_DIRECTORY}/manifests"
}

################ Main flow ########################################

if [[ $# -eq 0 ]]; then
    error "No parameters to the script, specify at least one like \"all\""
    exit 1
fi

while [[ $# -gt 0 ]]; do
    case "$1" in
    "all")
        deploy
        test_api
        test_staticroute
        get_manager_logs
        get_envoy_logs
        shift
        ;;

    "deploy")
        deploy
        shift
        ;;

    "test-api")
        test_api
        shift
        ;;

    "test-staticroute")
        test_staticroute
        shift
        ;;

    "delete")
        delete
        shift
        ;;

    *)
        # Non option argument
        break # Finish for loop
        ;;
    esac
done
