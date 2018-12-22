import React, { Component } from 'react'
import { Heading, Pane } from 'evergreen-ui'

import HistoryChart from './HistoryChart'

export default class ReportsOverTimePane extends Component {
  constructor (props) {
    super(props)

    this.state = {
      projectId: props.projectId || -1,
      reports: this.props.reportStore.state.reports
    }
  }

  async componentDidMount () {
    await this.props.reportStore.fetchReports()
    this.setState({
      reports: this.props.reportStore.state.reports
    })
  }

  async componentDidUpdate (prevProps) {
    if (prevProps.projectId === this.props.projectId &&
      !this.props.reportStore.isFetching) {
      const currentList = this.props.reportStore.state.reports
      const prevList = prevProps.reportStore.state.reports

      if (currentList.length === 0 && prevList.length > 0) {
        await this.props.reportStore.fetchReports()
        this.setState({
          reports: this.props.reportStore.state.reports
        })
      }
    }
  }

  render () {
    const reports = this.state.reports
    const hasReports = reports && reports.length > 0

    return (
      <Pane>
        <Pane display='flex' alignItems='center' marginTop={6}>
          <Heading size={500}>HISTORY</Heading>
        </Pane>
        {hasReports && <HistoryChart reports={reports} />}
      </Pane>
    )
  }
}
