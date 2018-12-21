import { Container } from 'unstated'

const projects = [
  {
    id: 11,
    name: 'Product User API - Service User',
    description: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi vel nibh interdum, rutrum mi non, elementum justo. Integer a massa maximus, facilisis sapien nec, pretium nunc. Donec maximus aliquam orci placerat venenatis. Mauris vel aliquet mauris. ',
    reports: 4,
    status: 'OK'
  },
  {
    id: 12,
    name: 'Project User API - Service Config',
    description: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi vel nibh interdum, rutrum mi non, elementum justo. Integer a massa maximus, facilisis sapien nec, pretium nunc. Donec maximus aliquam orci placerat venenatis. Mauris vel aliquet mauris. ',
    reports: 7,
    status: 'FAIL'
  },
  {
    id: 13,
    name: 'Component Event API - Service Ticket',
    description: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi vel nibh interdum, rutrum mi non, elementum justo. Integer a massa maximus, facilisis sapien nec, pretium nunc. Donec maximus aliquam orci placerat venenatis. Mauris vel aliquet mauris. ',
    reports: 12,
    status: 'OK'
  }
]

function getRandomInt (max) {
  return Math.floor(Math.random() * Math.floor(max))
}

async function getProjects (existing, sort) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      let data = existing
      if (!data || !data.length) {
        data = projects
      }

      if (sort) {
        data = data.reverse()
      }

      resolve(data)
    }, getRandomInt(800))
  })
}

async function createProject (name, description) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      const id = getRandomInt(10000)
      const newProject = {
        id,
        name,
        description,
        reports: 0,
        status: id > 5000 ? 'FAIL' : 'OK'
      }

      resolve(newProject)
    }, getRandomInt(800))
  })
}

export default class ProjectContainer extends Container {
  constructor (props) {
    super(props)
    this.state = {
      projects: [],
      isFetching: false
    }
  }

  async fetchProjects (sort) {
    this.setState({
      isFetching: true
    })

    try {
      const data = await getProjects(this.state.projects, sort)

      this.setState({
        projects: data,
        isFetching: false
      })
    } catch (err) {
      console.log('error: ', err)
    }
  }

  async createProject (name, desc) {
    this.setState(state => {
      return {
        ...this.state, isFetching: true
      }
    })

    try {
      const newProject = await createProject(name, desc)
      this.setState({
        projects: [...this.state.projects, newProject],
        isFetching: false
      })
    } catch (err) {
      console.log('error: ', err)
    }
  }
}
