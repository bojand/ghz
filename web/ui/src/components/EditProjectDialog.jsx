import React, { Component } from 'react'
import { Dialog, TextInputField, Textarea, Pane, Button } from 'evergreen-ui'

export default class EditProjectDialog extends Component {
  constructor (props) {
    super(props)
    this.state = {
      isShown: false,
      isLoading: false,
      name: '',
      description: '',
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
          }}
          onConfirm={async () => {
            if (this.state.name.trim() === '') {
              this.setState({ ...this.state, isInvalid: true })
              return
            }
            this.setState({ ...this.state, isLoading: true })
            await this.props.projectStore.createProject(this.state.name, this.state.description)
            this.setState({ ...this.state, isLoading: false, isShown: false })
          }}
          isConfirmLoading={this.state.isLoading}
          confirmLabel='Save'>
          <TextInputField
            required
            isInvalid={this.state.isInvalid}
            inputHeight={40}
            label='Name'
            placeholder='Name of the project'
            text={this.state.name}
            onChange={ev => this.onChangeText('name', ev.target.value)}
          />
          <Textarea
            label='Description'
            placeholder='Description of the project'
            text={this.state.description}
            onChange={ev => this.onChangeText('description', ev.target.value)}
          />

        </Dialog>

        <Button onClick={() => this.setState({ isShown: true })} marginLeft={14} iconBefore='plus' appearance='minimal' intent='none'>NEW</Button>
      </Pane>
    )
  }
}
