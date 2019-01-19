import { Container } from 'unstated'
import ky from 'ky'
import _ from 'lodash'
import { toaster } from 'evergreen-ui'

import { getAppRoot } from '../lib/common'

const api = ky.extend({ prefixUrl: getAppRoot() + '/api/' })

export default class ProjectContainer extends Container {
  constructor (props) {
    super(props)
    this.state = {
      totalProjects: 0,
      projects: [],
      isFetching: false,
      currentProject: {}
    }
  }

  async fetchProjects (order = 'desc', page = 0) {
    this.setState({
      isFetching: true
    })

    const searchParams = new URLSearchParams()

    if (order) {
      searchParams.append('order', order)
      searchParams.append('page', page)
    }

    try {
      const { data, total } = await api.get('projects', { searchParams }).json()

      this.setState({
        totalProjects: total,
        projects: data,
        isFetching: false
      })
    } catch (err) {
      toaster.danger(err.message)
      console.log('error: ', err)
    }
  }

  async createProject (name, description) {
    this.setState({
      isFetching: true
    })

    try {
      const newProject = await api.post('projects', { json: { name, description } }).json()
      this.setState({
        projects: [newProject, ...this.state.projects],
        totalProjects: this.state.totalProjects + 1,
        isFetching: false
      })
    } catch (err) {
      toaster.danger(err.message)
      console.log('error: ', err)
    }
  }

  async updateProject (id, name, description) {
    this.setState({
      isFetching: true
    })

    try {
      const newProject = await api.put(`projects/${id}`, { json: { name, description } }).json()

      const index = _.findIndex(this.state.projects, p => p.id.toString() === id.toString())

      let projects = this.state.projects
      if (index >= 0) {
        projects[index] = newProject
      }

      this.setState({
        projects,
        totalProjects: this.state.totalProjects + 1,
        isFetching: false,
        currentProject: newProject
      })
    } catch (err) {
      toaster.danger(err.message)
      console.log('error: ', err)
    }
  }

  async fetchProject (id) {
    this.setState({
      isFetching: true
    })

    try {
      const project = await api.get(`projects/${id}`).json()
      this.setState({
        currentProject: project,
        isFetching: false
      })
    } catch (err) {
      toaster.danger(err.message)
      console.log('error: ', err)
    }
  }

  async deleteProject (id) {
    this.setState({
      isFetching: true
    })

    try {
      await api.delete(`projects/${id}`).json()
      this.setState({
        isFetching: false
      })

      return true
    } catch (err) {
      toaster.danger(err.message)
      console.log('error: ', err)
    }
  }
}
