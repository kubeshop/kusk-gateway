# Privacy Policy

With the aim to improve the user experience, Kusk collects anonymous usage data.

You may [opt-out](#how-to-opt-out) if you'd prefer not to share any information.

The data collected is always anonymous, not traceable to the source, and only used in aggregate form. 

Telemetry collects and scrambles information about the host when the API server is bootstrapped for the first time. 

```json
{
   "anonymousId": "37c7dd3d2f0cd7eca8fdc5b606577278bf2a65e5da42fd4b809cfdf103583a98",
   "context": {
     "library": {
       "name": "analytics-go",
       "version": "3.0.0"
     }
   },
   "event": "kusk-cli",
   "integrations": {},
   "messageId": "c785d086-2d85-4d7a-9468-1da350822c95",
   "originalTimestamp": "2022-07-15T11:42:41.213006+08:00",
   "properties": {
     "event": "dashboard"
   },
   "receivedAt": "2022-07-15T03:42:42.691Z",
   "sentAt": "2022-07-15T03:42:41.215Z",
   "timestamp": "2022-07-15T03:42:42.689Z",
   "type": "track",
   "userId": "37c7dd3d2f0cd7eca8fdc5b606577278bf2a65e5da42fd4b809cfdf103583a98",
   "writeKey": "1t8VoI1wfqa43n0pYU01VZU2ZVDJKcQh"
 }
```

## **What We Collect**

The telemetry data we use in our metrics is limited to:

 - The number of CLI installations.
 - The number of unique CLI usages in a day.
 - The number of installations to a cluster.
 - The number of unique active cluster installations.
 - The number of people who disable telemetry.
 - The number of unique sessions in the UI.
 - The number of API, StaticRoute and EnvoyFleet creations

## How to opt out

### Helm Chart
To disable sending the anonymous analytics, provide the `analytics.enable: false` override during Helm chart installation or upgrade. See the <a href="https://github.com/kubeshop/helm-charts/blob/main/charts/kusk-gateway/values.yaml" target="_blank">Helm chart parameters</a> for more details about Helm chart configuration.

```
helm upgrade kusk-gateway kubeshop/kusk-gateway \
--install --namespace --create-namespace \
--set analytics.enabled=false \
...
```

### Kusk CLI
Set the following environment variable when running kusk commands
```
export ANALYTICS_ENABLED=false
```
or
```
ANALYTICS_ENABLED=false kusk install
```