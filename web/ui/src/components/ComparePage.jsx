import React, { Component } from 'react'
import { Pane } from 'evergreen-ui'
import { Provider, Subscribe } from 'unstated'

import ComparePane from './ComparePane'

import CompareContainer from '../containers/CompareContainer'

export default class ComparePage extends Component {
  render () {
    return (
      <Provider>
        <Subscribe to={[CompareContainer]}>
          {(compareStore) => (
            <Pane>
              <ComparePane
                compareStore={compareStore}
                reportId1={this.props.reportId1}
                reportId2={this.props.reportId2}
              />
            </Pane >
          )}
        </Subscribe>
      </Provider >
    )
  }
}
