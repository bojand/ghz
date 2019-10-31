import React, { Component } from 'react'
import { Pane, Heading, Button, Paragraph, toaster } from 'evergreen-ui'
import { toUpper } from 'lodash'
import { withRouter } from 'react-router-dom'

import EditProjectDialog from './EditProjectDialog'
import DeleteDialog from './DeleteDialog'
import StatusBadge from './StatusBadge'

class ProjectDetailPane extends Component {
  constructor (props) {
    super(props)

    this.state = {
      projectId: props.projectId || -1,
      editProjectVisible: false,
      deleteVisible: false
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

  async deleteProject () {
    this.setState({ deleteVisible: false })

    const currentProject = this.props.projectStore.state.currentProject
    const id = this.props.projectId
    const name = currentProject && currentProject.name ? currentProject.name : id

    const ok = await this.props.projectStore.deleteProject(id)
    if (ok) {
      toaster.success(`Project ${name} deleted.`)
      this.props.history.push('/projects')
    }
  }

  render () {
    const { state: { currentProject } } = this.props.projectStore

    if (!currentProject) {
      return (<Pane />)
    }

    return (
      <Pane>
        <Pane display='flex' marginBottom={10}>
          <Pane display='flex' alignItems='center' flex={1}>
            <StatusBadge status={currentProject.status} marginRight={8} />
            <Heading size={600}>{toUpper(currentProject.name)}</Heading>
            {this.state.editProjectVisible
              ? <EditProjectDialog
                projectStore={this.props.projectStore}
                project={currentProject}
                isShown={this.state.editProjectVisible}
                onDone={() => this.setState({ editProjectVisible: false })}
                /> : null}
            <Button
              onClick={() => this.setState({ editProjectVisible: !this.state.editProjectVisible })}
              marginLeft={14}
              iconBefore='edit'
              appearance='minimal'
              intent='none'
            >EDIT
            </Button>
          </Pane>
          <Pane display='flex'>
            {this.state.deleteVisible
              ? <DeleteDialog
                dataType='project'
                name={currentProject.name}
                isShown={this.state.deleteVisible}
                onConfirm={() => this.deleteProject()}
                onCancel={() => this.setState({ deleteVisible: !this.state.deleteVisible })}
                /> : null}
            <Button
              iconBefore='trash'
              appearance='minimal'
              intent='danger'
              onClick={() => this.setState({ deleteVisible: !this.state.deleteVisible })}
            >DELETE
            </Button>
          </Pane>
        </Pane>
        <Paragraph>{currentProject.description}</Paragraph>
      </Pane>
    )
  }
}

export default withRouter(ProjectDetailPane)
