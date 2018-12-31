import React, { Component } from 'react'
import { Pane, Text, Strong } from 'evergreen-ui'

export default class ReportDetailPane extends Component {
  componentDidMount () {
    this.props.infoStore.fetchInfo()
  }

  render () {
    const { state: { info } } = this.props.infoStore

    if (!info) {
      return (<Pane />)
    }

    return (
      <Pane>
        <Pane>
          <Text>
            <Strong>Version:</Strong> {info.version}
          </Text>
        </Pane>
        <Pane>
          <Text>
            <Strong>Uptime:</Strong> {info.uptime}
          </Text>
        </Pane>
      </Pane>
    )
  }
}
