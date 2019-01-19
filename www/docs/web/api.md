---
id: api
title: API
---

Reports are created in the system using the **Ingest API**. There are two endpoints for ingesting **raw JSON report** data into the system:

```sh
POST /api/ingest
```

This endpoint automatically creates a new project, ingests a report and assigns it to the project.

```sh
POST /api/projects/:id/ingest
```

Alternatively we can manually create a project ahead of time and then ingest reports specifically for an existing project using this endoint.

### Example

```sh
ghz -insecure \
    -proto ./greeter.proto \
    -call helloworld.Greeter.SayHello \
    -d '{"name": "Bob"}' \
    -tags '{"env": "staging", "created by":"Joe Developer"}' \
    -name 'Greeter SayHello' \
    -O json \
    0.0.0.0:50051 | http POST localhost:3000/api/projects/34/ingest
```
