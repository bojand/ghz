import React, { Component } from 'react'
import { Pane, Heading, Icon } from 'evergreen-ui'

import {
  getIconForStatus,
  getColorForStatus
} from '../lib/common'

export default class ProjectDetailPane extends Component {
  constructor (props) {
    super(props)

    this.state = {
      reportId: props.reportId || 0
    }
  }

  componentDidMount () {
    this.props.reportStore.fetchReport(this.props.reportId)
  }

  componentDidUpdate (prevProps) {
    if (!this.props.reportStore.state.isFetching &&
      (this.props.reportId !== prevProps.reportId)) {
      this.props.reportStore.fetchReport(this.props.reportId)
    }
  }

  render () {
    const { state: { currentReport } } = this.props.reportStore

    if (!currentReport) {
      return (<Pane />)
    }
    return (
      <Pane>
        <Pane display='flex' alignItems='center' marginTop={6} marginBottom={10}>
          <Icon
            marginRight={16}
            icon={getIconForStatus(currentReport.status)}
            color={getColorForStatus(currentReport.status)} />
          <Heading size={500}>
            {currentReport.name || `REPORT: ${currentReport.id}`}
          </Heading>
        </Pane>
      </Pane>
    )
  }
}
