import React, { Component } from 'react'
import { Pane } from 'evergreen-ui'
import { Provider, Subscribe } from 'unstated'

import ReportList from './ReportList'

import ReportContainer from '../containers/ReportContainer'

export default class ReportsPage extends Component {
  render () {
    return (
      <Provider>
        <Subscribe to={[ReportContainer]}>
          {(reportStore) => (
            <Pane>
              <ReportList reportStore={reportStore} />
            </Pane>
          )}
        </Subscribe>
      </Provider>
    )
  }
}
