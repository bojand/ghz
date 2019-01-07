import React, { Component } from 'react'
import { Pane } from 'evergreen-ui'
import { Provider, Subscribe } from 'unstated'

import ReportDetailPane from './ReportDetailPane'

import CompareContainer from '../containers/CompareContainer'

export default class ReportDetailPage extends Component {
  render () {
    return (
      <Provider>
        <Subscribe to={[CompareContainer]}>
          {compareStore => (
            <Pane>
              <Pane marginBottom={24}>
                <ReportDetailPane
                  compareStore={compareStore}
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
