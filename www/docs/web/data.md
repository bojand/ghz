---
id: data
title: Data
---

The main data components of the application are **Projects** and **Reports**.

A Report represent a result of a single **ghz** test run. Reports can be grouped together into **Projects** to be tracked, analyzed and compared over time. Thus it is important to group reports together that make sense, normally tests against a single same gRPC call.


For example if we are testing `helloworld.Greeter.SayHello` call, we may have a separate project for development, staging and production deployment environments for `helloworld.Greeter.SayHello` test only. For example our staging environment project may be called `"helloworld.Greeter.SayHello - staging"`.

We would keep test results against this gRPC call for each deployment environment grouped together respectively and therefore be able to track change in performance over time and compare different environments as our service changes.

### Status

Each Report and Project has a status associated with it. A Status can be either `OK` or `FAIL`. If the test result had any errors in it then its status will be `FAIL`. Similarly a projects status always reflects the status of the latest report created for it.
