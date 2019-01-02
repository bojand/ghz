import React, { Component } from 'react'
import { Pane, Heading, Button, Paragraph } from 'evergreen-ui'
import { toUpper } from 'lodash'

import EditProjectDialog from './EditProjectDialog'
import StatusBadge from './StatusBadge'

export default class ProjectDetailPane extends Component {
  constructor (props) {
    super(props)

    this.state = {
      projectId: props.projectId || -1,
      editProjectVisible: false
    }
  }

  componentDidMount () {
    this.props.projectStore.fetchProject(this.props.projectId)
  }

  componentDidUpdate (prevProps) {
    if (!this.props.projectStore.state.isFetching &&
      (this.props.projectId !== prevProps.projectId)) {
      this.props.projectStore.fetchProject(this.props.projectId)
    }
  }

  render () {
    const { state: { currentProject } } = this.props.projectStore

    if (!currentProject) {
      return (<Pane />)
    }
    return (
      <Pane>
        <Pane display='flex' alignItems='center' marginBottom={10}>
          <StatusBadge status={currentProject.status} marginRight={8} />
          <Heading size={500}>{toUpper(currentProject.name)}</Heading>
          {this.state.editProjectVisible
            ? <EditProjectDialog
              projectStore={this.props.projectStore}
              project={currentProject}
              isShown={this.state.editProjectVisible}
              onDone={() => this.setState({ editProjectVisible: false })}
            /> : null
          }
          <Button onClick={() => this.setState({ editProjectVisible: !this.state.editProjectVisible })} marginLeft={14} iconBefore='edit' appearance='minimal' intent='none'>EDIT</Button>
        </Pane>
        <Paragraph>{currentProject.description}</Paragraph>
      </Pane>
    )
  }
}
