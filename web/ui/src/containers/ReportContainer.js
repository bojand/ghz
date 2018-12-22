import { Container } from 'unstated'

const reports = [{
  id: 10,
  date: new Date().toISOString(),
  projectId: 11,
  count: 100,
  total: 172959832,
  average: 31877742,
  averageStatus: 'down_better',
  fastest: 25404280,
  fastestStatus: 'up_worse',
  slowest: 62984994,
  slowestStatus: 'up_worse',
  rps: 1156.34,
  rpsStatus: 'up_better',
  errorDistribution: {
    'rpc error: code = Internal desc = Internal error.': 5,
    'rpc error: code = PermissionDenied desc = Permission denied.': 4
  },
  statusCodeDistribution: {
    'OK': 191
  },
  status: 'FAIL'
}, {
  id: 20,
  date: new Date().toISOString(),
  projectId: 12,
  count: 200,
  total: 272959832,
  average: 41877742,
  averageStatus: 'up_worse',
  fastest: 15404280,
  fastestStatus: 'down_better',
  slowest: 72984994,
  slowestStatus: 'up_worse',
  rps: 2156.34,
  rpsStatus: 'up_better',
  errorDistribution: {
    'rpc error: code = Internal desc = Internal error.': 5,
    'rpc error: code = PermissionDenied desc = Permission denied.': 4
  },
  statusCodeDistribution: {
    'OK': 191
  },
  status: 'OK'
}]

function getRandomInt (max) {
  return Math.floor(Math.random() * Math.floor(max))
}

async function getReports (existing, sort) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      let data = existing
      if (!data || !data.length) {
        data = reports
      }

      if (sort) {
        data = data.reverse()
      }

      resolve(data)
    }, getRandomInt(800))
  })
}

export default class ReportContainer extends Container {
  constructor (props) {
    super(props)
    this.state = {
      reports: [],
      isFetching: false
    }
  }

  async fetchReports (sort) {
    this.setState({
      isFetching: true
    })

    try {
      const data = await getReports(this.state.reports, sort)

      this.setState({
        reports: data,
        isFetching: false
      })
    } catch (err) {
      console.log('error: ', err)
    }
  }
}
