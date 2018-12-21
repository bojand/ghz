import React from 'react'
import { Pane, Icon, Text, Tab } from 'evergreen-ui'

export default class TopBar extends React.PureComponent {
  render () {
    return (
      <Pane display='flex' paddingBottom={8} borderBottom>
        <Pane marginLeft={8} flex={1} alignItems='center' display='flex'>
          <Tab height={36} paddingX={14}><Icon icon='control' marginRight={12} /><Text size={400}>PROJECTS</Text></Tab>
          <Tab height={36} paddingX={14}><Icon icon='dashboard' marginRight={12} /><Text size={400}>REPORTS</Text></Tab>
        </Pane>
        <Pane>
          <Tab height={36} paddingX={14}><Icon icon='manual' marginRight={12} /><Text size={400}>DOCS</Text></Tab>
          <Tab marginRight={14} height={36} paddingX={14}><Icon icon='info-sign' marginRight={12} /><Text size={400}>ABOUT</Text></Tab>
        </Pane>
      </Pane>
    )
  }
}
