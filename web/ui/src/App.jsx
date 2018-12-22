import React, { Component } from 'react'

import TopBar from './components/TopBar'
import ProjectPage from './components/ProjectPage'

export default class App extends Component {
  render () {
    return (
      <div>
        <TopBar />
        <ProjectPage paddingX={24} paddingY={14} marginTop={6} />
      </div >
    )
  }
}
