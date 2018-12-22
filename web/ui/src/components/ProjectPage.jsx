import React, { Component } from 'react'
import { Pane } from 'evergreen-ui'
import { Provider, Subscribe } from 'unstated'

import TopBar from './TopBar'
import ProjectList from './ProjectList'

import ProjectContainer from '../containers/ProjectContainer'

export default class ProjectPAge extends Component {
  render () {
    return (
      <Provider>
        <Subscribe to={[ProjectContainer]}>
          {(projectStore) => (
            <Pane>
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
