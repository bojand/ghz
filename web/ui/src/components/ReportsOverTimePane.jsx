import React, { Component } from 'react'
import { Heading, Pane } from 'evergreen-ui'

import HistoryChart from './HistoryChart'

export default class ReportsOverTimePane extends Component {
  constructor (props) {
    super(props)

    this.state = {
      projectId: props.projectId || 0,
      reports: this.props.reportStore.state.reports
    }
  }

  async componentDidMount () {
    await this.props.reportStore.fetchReports('desc', 'date', 0, this.state.projectId)
  }

  async componentDidUpdate (prevProps) {
    if (prevProps.projectId === this.props.projectId) {
      const currentList = this.props.reportStore.state.reports
      const prevList = prevProps.reportStore.state.reports

      if (!this.props.reportStore.state.isFetching) {
        if ((currentList.length === 0 && prevList.length > 0) || (prevList.length > currentList.length)) {
          await this.props.reportStore.fetchReports(
            'desc', 'date', 0, this.props.projectId
          )
        }
      }
    }
  }

  render () {
    const reports = this.props.reportStore.state.reports
    const hasReports = reports && reports.length > 0

    if (!hasReports) {
      return (<Pane />)
    }

    return (
      <Pane marginTop={24} marginBottom={24}>
        <Pane display='flex' alignItems='center' marginTop={6}>
          <Heading size={600}>HISTORY</Heading>
        </Pane>
        <Pane paddingX={20} paddingTop={20}>
          <HistoryChart
            reports={reports}
            projectId={this.state.projectId}
          />
        </Pane>
      </Pane>
    )
  }
}
