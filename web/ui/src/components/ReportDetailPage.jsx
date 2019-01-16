import React, { Component } from 'react'
import { Pane } from 'evergreen-ui'
import { Provider, Subscribe } from 'unstated'

import ReportDetailPane from './ReportDetailPane'

import CompareContainer from '../containers/CompareContainer'
import ReportContainer from '../containers/ReportContainer'

export default class ReportDetailPage extends Component {
  render () {
    return (
      <Provider>
        <Subscribe to={[CompareContainer, ReportContainer]}>
          {(compareStore, reportStore) => (
            <Pane>
              <Pane marginBottom={24}>
                <ReportDetailPane
                  compareStore={compareStore}
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
