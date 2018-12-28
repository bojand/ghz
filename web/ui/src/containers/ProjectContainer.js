import { Container } from 'unstated'
import ky from 'ky'
import { toaster } from 'evergreen-ui'

const api = ky.extend({ prefixUrl: 'http://localhost:3000/api/' })

export default class ProjectContainer extends Container {
  constructor (props) {
    super(props)
    this.state = {
      projects: [],
      isFetching: false,
      currentProject: {}
    }
  }

  async fetchProjects (sort) {
    this.setState({
      isFetching: true
    })

    const searchParams = new URLSearchParams()

    if (sort) {
      searchParams.append('sort', sort)
    }

    try {
      const { data } = await api.get('projects', { searchParams }).json()

      this.setState({
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
        isFetching: false
      })
    } catch (err) {
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
      console.log('error: ', err)
    }
  }
}
