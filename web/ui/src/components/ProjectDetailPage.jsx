import React, { Component } from 'react'
import { Pane } from 'evergreen-ui'
import { Provider, Subscribe } from 'unstated'

import ReportList from './ReportList'
import ProjectDetailPane from './ProjectDetailPane'
import ReportsOverTimePane from './ReportsOverTimePane'

import ReportContainer from '../containers/ReportContainer'
import ProjectContainer from '../containers/ProjectContainer'

export default class ProjectDetailPage extends Component {
  render () {
    return (
      <Provider>
        <Subscribe to={[ReportContainer, ProjectContainer]}>
          {(reportStore, projectStore) => (
            <Pane>
              <Pane marginBottom={24}>
                <ProjectDetailPane
                  projectStore={projectStore}
                  projectId={this.props.projectId}
                />
              </Pane>

              <Pane marginBottom={24}>
                <ReportsOverTimePane
                  reportStore={reportStore}
                  projectId={this.props.projectId}
                />
              </Pane>

              <Pane>
                <ReportList
                  reportStore={reportStore}
                  projectId={this.props.projectId}
                />
              </Pane>
            </Pane>
          )}
        </Subscribe>
      </Provider>
    )
  }
}
