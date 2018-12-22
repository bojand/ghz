import React, { Component } from 'react'
import { Dialog, TextInputField, Textarea, Pane } from 'evergreen-ui'

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
    return (
      <Pane>
        <Dialog
          isShown={this.state.isShown}
          title='New Project'
          onCloseComplete={() => {
            this.setState({ ...this.state, isShown: false, isLoading: false })
            if (typeof this.props.onDone === 'function') {
              this.props.onDone()
            }
          }}
          onConfirm={async () => {
            if (this.state.name.trim() === '') {
              this.setState({ ...this.state, isInvalid: true })
              return
            }
            this.setState({ ...this.state, isLoading: true })
            await this.props.projectStore.createProject(this.state.name, this.state.description)
            this.setState({ ...this.state, isLoading: false, isShown: false })
            console.log(this.props.onDone)
            if (typeof this.props.onDone === 'function') {
              this.props.onDone()
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
          <Textarea
            label='Description'
            placeholder='Description of the project'
            value={this.state.description}
            onChange={ev => this.onChangeText('description', ev.target.value)}
          />

        </Dialog>
      </Pane>
    )
  }
}
