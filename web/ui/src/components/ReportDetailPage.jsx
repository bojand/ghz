import React, { Component } from 'react'
import { Pane } from 'evergreen-ui'
import { Provider, Subscribe } from 'unstated'

import ReportDetailPane from './ReportDetailPane'

import ReportContainer from '../containers/ReportContainer'

export default class ProjectDetailPage extends Component {
  render () {
    return (
      <Provider>
        <Subscribe to={[ReportContainer]}>
          {reportStore => (
            <Pane>
              <Pane marginBottom={24}>
                <ReportDetailPane
                  reportStore={reportStore}
                  reportId={this.props.projectId}
                />
              </Pane>
            </Pane>
          )}
        </Subscribe>
      </Provider >
    )
  }
}
