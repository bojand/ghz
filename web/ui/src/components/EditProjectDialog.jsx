import React, { Component } from 'react'
import { Dialog, TextInputField, Textarea, Pane, Label } from 'evergreen-ui'

export default class EditProjectDialog extends Component {
  constructor (props) {
    super(props)

    this.state = {
      isShown: props.isShown,
      isLoading: false,
      name: props.project ? props.project.name || '' : '',
      description: props.project ? props.project.description || '' : '',
      isInvalid: false
    }
  }

  onChangeText (key, value) {
    this.setState({
      ...this.state,
      [key]: value
    })
  }

  render () {
    const editId = this.props.project && this.props.project.id
    return (
      <Pane>
        <Dialog
          isShown={this.state.isShown}
          title={editId ? `Edit Project (ID: ${editId})` : 'New Project'}
          onCloseComplete={() => {
            this.setState({ isShown: false, isLoading: false })
            if (typeof this.props.onDone === 'function') {
              this.props.onDone()
            }
          }}
          onConfirm={async () => {
            if (this.state.name.trim() === '') {
              this.setState({ isInvalid: true })
              return
            }
            this.setState({ isLoading: true })

            let newProject = null

            if (editId) {
              newProject = await this.props.projectStore.updateProject(
                editId, this.state.name, this.state.description)
            } else {
              newProject = await this.props.projectStore.createProject(this.state.name, this.state.description)
            }

            this.setState({ ...this.state, isLoading: false, isShown: false })
            if (typeof this.props.onDone === 'function') {
              this.props.onDone(newProject)
            }
          }}
          isConfirmLoading={this.state.isLoading}
          confirmLabel='Save'>
          <TextInputField
            required
            isInvalid={this.state.isInvalid}
            inputHeight={40}
            label='Name'
            placeholder='Name of the project'
            value={this.state.name}
            onChange={ev => this.onChangeText('name', ev.target.value)}
          />
          <Label
            htmlFor='projectDescriptionTextarea'
            marginBottom={4}
            display='block'
          >
            Description
          </Label>
          <Textarea
            id='projectDescriptionTextarea'
            placeholder='Description of the project'
            value={this.state.description}
            onChange={ev => this.onChangeText('description', ev.target.value)}
          />

        </Dialog>
      </Pane>
    )
  }
}
