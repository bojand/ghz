import React, { Component } from 'react'
import { Pane, Heading, Paragraph, Pre, Code, Link, Strong } from 'evergreen-ui'

export default class DocsComponent extends Component {
  render () {
    return (
      <Pane marginTop={8} marginBottom={24}>
        <Heading size={600}>
          Documentation
        </Heading>

        <Heading size={600} marginTop={16} marginBottom={8}>Introduction</Heading>
        <Paragraph>
          <Strong>ghz-web</Strong> is a complementary web application that can be used to store and view <Link href='https://ghz.sh'>ghz</Link> test reports.
        </Paragraph>

        <Heading size={600} marginTop={16} marginBottom={8}>Data</Heading>
        <Paragraph>
          The main data components of the application are <Strong>Projects</Strong> and <Strong>Reports</Strong>.
          A Report represent a result of a single <Strong>ghz</Strong> test run. Reports can be grouped together into <Strong>Projects</Strong> to be tracked, analyzed and compared over time.
          Thus it is important to group reports together that make sense, normally tests against a single same gRPC call.
          For example if we are testing <Code size={300}>helloworld.Greeter.SayHello</Code> call, we may have a separate project for development, staging and production deployment environments for <Code size={300}>helloworld.Greeter.SayHello</Code> test only.
          For example our staging environment project may be called <Code size={300}>"helloworld.Greeter.SayHello - staging"</Code>.
          We would keep test results against this gRPC call for each deployment environment grouped together respectively and therefore be able to track change in performance over time and compare different environments as our service changes.
        </Paragraph>

        <Heading size={600} marginTop={16} marginBottom={8}>Ingest API</Heading>
        <Paragraph>
          Reports are created in the system using the <Strong>Ingest API</Strong>. There are two endpoints for ingesting <Strong>raw JSON report</Strong> data into the system:
        </Paragraph>
        <Pane background='tint2' marginTop={12} marginBottom={12} maxWidth={640}>
          <Pre fontFamily='monospace' padding={12}>
            POST /api/ingest
          </Pre>
        </Pane>
        <Paragraph>
          When used this endpoing automatically created a new project, ingests and report and assigns it to the project.
        </Paragraph>
        <Pane background='tint2' marginTop={12} marginBottom={12} maxWidth={640}>
          <Pre fontFamily='monospace' padding={12}>
            POST /api/projects/:id/ingest
          </Pre>
        </Pane>
        <Paragraph>
          Alternatively we can manually create a project ahead of time and then ingest reports specifically for an existing project using this endoint.
        </Paragraph>
        <Heading marginTop={16} marginBottom={8} size={400}>Exmaple</Heading>
        <Paragraph>
          Example of a test run performed against a service running locally, and then ingested using <Link href='https://httpie.org'>HTTPie</Link>, with ghz-web also running locally.
          You would substitute the host and port to match your installation.
          In this case we have a project already created, and its <Code size={300}>ID</Code> is <Code size={300}>34</Code>.
        </Paragraph>
        <Pane background='tint2' marginTop={12} marginBottom={12} maxWidth={640}>
          <Pre fontFamily='monospace' padding={12}>
            {`ghz -insecure \\
    -proto ./testdata/greeter.proto \\
    -call helloworld.Greeter.SayHello \\
    -d '{"name": "Bob"}' \\
    -tags '{"env": "staging", "created by":"Joe Developer"}' \\
    -O pretty \\
    -name 'Greeter SayHello' \\
    0.0.0.0:50051 | http POST localhost:3000/api/projects/34/ingest`}
          </Pre>
        </Pane>

        <Heading size={600} marginTop={16} marginBottom={8}>Status</Heading>
        <Paragraph>
          Each Report and Project has a status associated with it. A Status can be either <Code size={300}>OK</Code> or <Code size={300}>FAIL</Code>. If the test result had any errors in it then its status will be <Code size={300}>FAIL</Code>. Similarly a projects status always reflects the status of the latest report created for it.
        </Paragraph>
      </Pane>
    )
  }
}
