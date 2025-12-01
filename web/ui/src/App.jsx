import React, { Component } from 'react'
import { Pane, Tab, Icon, Text } from 'evergreen-ui'
import { BrowserRouter as Router, Route, Link as RouterLink, Switch, useLocation } from 'react-router-dom'
import { Provider, Subscribe } from 'unstated'

import ProjectListPage from './components/ProjectListPage'
import ProjectDetailPage from './components/ProjectDetailPage'
import ReportPage from './components/ReportPage'
import ReportDetailPage from './components/ReportDetailPage'
import Footer from './components/Footer'
import InfoComponent from './components/InfoComponent'
import ComparePage from './components/ComparePage'
import RunTestPage from './components/RunTestPage'

import InfoContainer from './containers/InfoContainer'

export default class App extends Component {
  render () {
    return (
      <Router>
        <div>
          <Pane display='flex' paddingY={12} borderBottom>
            <Pane flex={1} alignItems='center' display='flex' marginLeft={8}>
              <NavTabs />
            </Pane>
          </Pane>
          <Switch>
            <Route exact path='/' component={Projects} />
            <Route path='/projects/:projectId/run' component={RunInProject} />
            <Route path='/projects/:projectId' component={Projects} />
            <Route path='/projects' component={Projects} />
            <Route path='/reports/:reportId' component={Reports} />
            <Route path='/compare/:reportId1/:reportId2' component={Compare} />
            <Route path='/reports' component={Reports} />
            <Route path='/run' component={Run} />
            <Route path='/about' component={Info} />
          </Switch>
          <Footer />
        </div>
      </Router>
    )
  }
}

function Projects ({ match }) {
  return (
    <Pane minHeight={600} paddingX={24} paddingY={10} marginTop={6}>
      {match.params.projectId
        ? <ProjectDetailPage projectId={match.params.projectId} />
        : <ProjectListPage />
      }
    </Pane>
  )
}

function Reports ({ match }) {
  return (
    <Pane minHeight={600} paddingX={24} paddingY={10} marginTop={6}>
      {match.params.reportId
        ? <ReportDetailPage projectId={match.params.reportId} />
        : <ReportPage />
      }
    </Pane>
  )
}

function Compare ({ match }) {
  return (
    <Pane minHeight={600} paddingX={24} paddingY={10} marginTop={6}>
      <ComparePage reportId1={match.params.reportId1} reportId2={match.params.reportId2} />
    </Pane>
  )
}

function Run () {
  return (
    <Pane minHeight={600} paddingX={24} paddingY={10} marginTop={6}>
      <RunTestPage />
    </Pane>
  )
}

function RunInProject ({ match }) {
  return (
    <Pane minHeight={600} paddingX={24} paddingY={10} marginTop={6}>
      <RunTestPage projectId={match.params.projectId} />
    </Pane>
  )
}

function Info () {
  return (
    <Pane minHeight={600} paddingX={24} paddingY={10} marginTop={6}>
      <Provider>
        <Subscribe to={[InfoContainer]}>
          {infoStore => (
            <InfoComponent infoStore={infoStore} />
          )}
        </Subscribe>
      </Provider>
    </Pane>
  )
}

const NavTabs = () => {
  const location = useLocation()
  const tabs = [
    { to: '/projects', text: 'PROJECTS', icon: 'control' },
    { to: '/reports', text: 'REPORTS', icon: 'dashboard' },
    { to: '/run', text: 'RUN', icon: 'play' }
  ]
  return (
    <>
      {tabs.map(t => {
        const active = location.pathname.startsWith(t.to)
        return (
          <RouterLink key={t.to} to={t.to} style={{ textDecoration: 'none' }}>
            <Tab height={36} paddingX={14} isSelected={active} aria-current={active ? 'page' : undefined}>
              <Icon icon={t.icon} marginRight={12} /><Text size={400}>{t.text}</Text>
            </Tab>
          </RouterLink>
        )
      })}
    </>
  )
}
