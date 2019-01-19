import React, { Component } from 'react'
import { Dialog, Pane, Paragraph, TextInputField, Strong } from 'evergreen-ui'

export default class DeleteDialog extends Component {
  constructor (props) {
    super(props)

    this.state = {
      isShown: props.isShown,
      name: '',
      description: props.project ? props.project.description || '' : '',
      isInvalid: true
    }
  }

  onChangeText (key, value) {
    if (key === 'name') {
      this.setState({
        name: value,
        isInvalid: value.trim() !== this.props.name
      })
    }
  }

  render () {
    return (
      <Pane>
        <Dialog
          isShown={this.state.isShown}
          title={`Delete ${this.props.dataType} ${this.props.name}?`}
          intent='danger'
          isConfirmDisabled={this.state.isInvalid}
          onCancel={this.props.onCancel}
          onConfirm={() => {
            if (this.state.name.trim() !== this.props.name) {
              this.setState({ isInvalid: true })
              return
            }

            this.setState({ isShown: false })
            if (typeof this.props.onConfirm === 'function') {
              this.props.onConfirm()
            }
          }}
          confirmLabel='Delete'>
          <Paragraph>{`This will delete ${this.props.dataType} ${this.props.name} and all associated data.
            This action cannot be reveresed. Are you sure you want to proceed?
            If so type in the name of the ${this.props.dataType}: `} <Strong>{this.props.name}</Strong>
          </Paragraph>
          <TextInputField
            marginTop={16}
            required
            isInvalid={this.state.isInvalid}
            inputHeight={40}
            label='Name'
            placeholder={`Name of the ${this.props.dataType}`}
            value={this.state.name}
            onChange={ev => this.onChangeText('name', ev.target.value)}
          />
        </Dialog>
      </Pane>
    )
  }
}
