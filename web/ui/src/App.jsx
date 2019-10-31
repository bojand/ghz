import React, { Component } from 'react'
import { Pane, Tab, Icon, Text } from 'evergreen-ui'
import { BrowserRouter as Router, Route, Link as RouterLink, Switch } from 'react-router-dom'
import { Provider, Subscribe } from 'unstated'

import ProjectListPage from './components/ProjectListPage'
import ProjectDetailPage from './components/ProjectDetailPage'
import ReportPage from './components/ReportPage'
import ReportDetailPage from './components/ReportDetailPage'
import Footer from './components/Footer'
import InfoComponent from './components/InfoComponent'
import ComparePage from './components/ComparePage'

import InfoContainer from './containers/InfoContainer'

export default class App extends Component {
  render () {
    return (
      <Router>
        <div>
          <Pane display='flex' paddingY={12} borderBottom>
            <Pane flex={1} alignItems='center' display='flex' marginLeft={8}>
              <TabLink to='/projects' linkText='PROJECTS' icon='control' />
              <TabLink to='/reports' linkText='REPORTS' icon='dashboard' />
            </Pane>
          </Pane>
          <Switch>
            <Route exact path='/' component={Projects} />
            <Route path='/projects/:projectId' component={Projects} />
            <Route path='/projects' component={Projects} />
            <Route path='/reports/:reportId' component={Reports} />
            <Route path='/compare/:reportId1/:reportId2' component={Compare} />
            <Route path='/reports' component={Reports} />
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

const TabLink = ({ to, linkText, icon, ...rest }) => (
  <Route
    path={to}
    children={() => (
      <RouterLink to={to} {...rest} style={{ textDecoration: 'none' }}>
        <Tab height={36} paddingX={14}><Icon icon={icon} marginRight={12} /><Text size={400}>{linkText}</Text></Tab>
      </RouterLink>
    )
    }
  />
)
