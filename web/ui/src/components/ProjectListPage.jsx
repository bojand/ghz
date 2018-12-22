import React, { Component } from 'react'
import { Pane } from 'evergreen-ui'
import { Provider, Subscribe } from 'unstated'

import ProjectList from './ProjectList'

import ProjectContainer from '../containers/ProjectContainer'

export default class ProjectListPage extends Component {
  render () {
    return (
      <Provider>
        <Subscribe to={[ProjectContainer]}>
          {(projectStore) => (
            <Pane>
              <ProjectList projectStore={projectStore} />
            </Pane>
          )}
        </Subscribe>
      </Provider >
    )
  }
}
