import React, { Component } from 'react'
import { Pane, Strong, Link, Paragraph } from 'evergreen-ui'
import { Link as RouterLink } from 'react-router-dom'
import GitHubIcon from './GitHubIcon'

export default class Footer extends Component {
  render () {
    return (
      <Pane
        borderTop
        background='tint2'
        alignItems='center'
        justifyContent='center'
        display='flex'
        flexDirection='column'
        minHeight={120}>
        <Pane maxHeight={30} maxWidth={30} marginBottom={10}>
          <Link href='https://github.com/bojand/ghz'>
            <GitHubIcon />
          </Link>
        </Pane>

        <Paragraph>
          <Strong>ghz</Strong> The source code is licensed <Link href='https://opensource.org/licenses/Apache-2.0'>Apache-2.0</Link>.
        </Paragraph>
        <Paragraph>
          <RouterLink to='/about'>
            About
          </RouterLink>
        </Paragraph>
      </Pane>
    )
  }
}
