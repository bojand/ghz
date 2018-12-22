import React, { Component } from 'react'
import { Pane, Tab, Icon, Text } from 'evergreen-ui'
import { BrowserRouter as Router, Route, Link } from 'react-router-dom'

import ProjectPage from './components/ProjectPage'
import ReportPage from './components/ReportPage'

export default class App extends Component {
  render () {
    return (
      <Router>
        <div>
          <Pane display='flex' paddingBottom={8} marginLeft={8} marginRight={8} borderBottom>
            <Pane flex={1} alignItems='center' display='flex'>
              <TabLink to='/projects' linkText='PROJECTS' icon='control' />
              <TabLink to='/reports' linkText='REPORTS' icon='dashboard' />
            </Pane>
            <Pane>
              <Tab height={36} paddingX={14}><Icon icon='manual' marginRight={12} /><Text size={400}>DOCS</Text></Tab>
              <Tab height={36} paddingX={14}><Icon icon='info-sign' marginRight={12} /><Text size={400}>ABOUT</Text></Tab>
            </Pane>
          </Pane>
          <Route exact path='/' component={Projects} />
          <Route path='/projects' component={Projects} />
          <Route path='/reports' component={Reports} />
        </div>
      </Router>
    )
  }
}

function Projects (props) {
  return (
    <Pane paddingX={24} paddingY={10} marginTop={6} >
      <ProjectPage />
    </Pane>
  )
}

function Reports (props) {
  return (
    <Pane paddingX={24} paddingY={10} marginTop={6} >
      <ReportPage />
    </Pane>
  )
}

const TabLink = ({ to, linkText, icon, ...rest }) => (
  <Route
    path={to}
    children={({ match }) => (
      <Link to={to} {...rest} >
        <Tab height={36} paddingX={14}><Icon icon={icon} marginRight={12} /><Text size={400}>{linkText}</Text></Tab>
      </Link>
    )
    }
  />
)
