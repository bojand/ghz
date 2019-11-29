---
id: extras
title: Extras
---

## Grafana dashboards

For convenience we include prebuilt [Grafana](http://grafana.com/) dashboards for [summary](/extras/influx-summary-grafana-dashboard.json) and [details](/extras/influx-details-grafana-dashboard.json).

#### Summary Grafana Dashboard

<div align="center">
	<br>
	<img src="/img/influx-summary-grafana-dashboard.png" alt="Summary Grafana Dashboard">
	<br>
</div>

#### Details Grafana Dashboard:

<div align="center">
	<br>
	<img src="/img/influx-details-grafana-dashboard.png" alt="Details Grafana Dashboard">
	<br>
</div>

## Prototool

`ghz` can be used with [Prototool](https://github.com/uber/prototool) using the [`descriptor-set`](https://github.com/uber/prototool/tree/dev/docs#prototool-descriptor-set) command:

```
ghz -protoset $(prototool descriptor-set --include-imports --tmp) ...
```
