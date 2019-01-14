import React, { Component } from 'react'
import { Pane, Heading, Pre } from 'evergreen-ui'

import { pretty } from '../lib/common'

export default class OptionsPane extends Component {
  constructor (props) {
    super(props)

    this.state = {
      reportId: props.reportId || 0
    }
  }

  componentDidMount () {
    this.props.optionsStore.fetchOptions(this.props.reportId)
  }

  componentDidUpdate (prevProps) {
    if (!this.props.optionsStore.state.isFetching &&
      (this.props.reportId !== prevProps.reportId)) {
      this.props.optionsStore.fetchOptions(this.props.reportId)
    }
  }

  render () {
    const { state: { options } } = this.props.optionsStore

    if (!options || !options.call) {
      return (<Pane />)
    }

    return (
      <Pane>
        <Heading size={600}>
          Options
        </Heading>
        <Pane background='tint2' marginTop={16}>
          <Pre fontFamily='monospace' paddingX={16} paddingY={16}>
            {pretty(options)}
          </Pre>
        </Pane>
      </Pane>
    )
  }
}
