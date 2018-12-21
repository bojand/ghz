import React, { Component } from 'react'
import { Pane } from 'evergreen-ui'
import { Provider, Subscribe } from 'unstated'

import TopBar from './components/TopBar'
import ProjectList from './components/ProjectList'

import ProjectContainer from './containers/ProjectContainer'

export default class App extends Component {
  render () {
    return (
      <Provider>
        <Subscribe to={[ProjectContainer]}>
          {(projectStore) => (
            <Pane>
              <TopBar />
              <Pane paddingX={24} paddingY={14} marginTop={6}>
                <ProjectList projectStore={projectStore} />
              </Pane>
            </Pane>
          )}
        </Subscribe>
      </Provider >
    )
  }
}
