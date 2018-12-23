import React, { Component } from 'react'
import { Pane, Strong, Link, Paragraph } from 'evergreen-ui'
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
          <Strong>ghz</Strong> by <Link href='https://bojand.github.io'>Bojan D</Link>.
              The source code is licensed <Link href='http://opensource.org/licenses/mit-license.php'>MIT</Link>.
        </Paragraph>
      </Pane>
    )
  }
}
