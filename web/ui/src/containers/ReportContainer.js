import { Container } from 'unstated'
import _ from 'lodash'

const date1 = new Date()
const date2 = new Date()
const date3 = new Date()
const date4 = new Date()

date2.setMonth(date1.getMonth() - 1)
date3.setMonth(date2.getMonth() - 1)
date4.setMonth(date3.getMonth() - 1)

const reports = [{
  id: 10,
  date: date1.toISOString(),
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
  status: 'FAIL',
  'options': {
    'call': 'helloworld.Greeter.SayHello',
    'proto': '../../testdata/greeter.proto',
    'host': '0.0.0.0:50051',
    'n': 200,
    'c': 50,
    'timeout': 20000000000,
    'dialTimeout': 10000000000,
    'data': {
      'name': 'Joe'
    },
    'binary': false,
    'insecure': true,
    'CPUs': 8
  }
}, {
  id: 20,
  date: date2.toISOString(),
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
  status: 'FAIL',
  'options': {
    'call': 'helloworld.Greeter.SayHello',
    'proto': '../../testdata/greeter.proto',
    'host': '0.0.0.0:50051',
    'n': 200,
    'c': 50,
    'timeout': 20000000000,
    'dialTimeout': 10000000000,
    'data': {
      'name': 'Joe'
    },
    'binary': false,
    'insecure': true,
    'CPUs': 8
  }
}, {
  id: 22,
  date: date3.toISOString(),
  projectId: 12,
  count: 300,
  total: 472959832,
  average: 51877742,
  averageStatus: 'up_worse',
  fastest: 59404280,
  fastestStatus: 'up_worse',
  slowest: 58984994,
  slowestStatus: 'up_worse',
  rps: 1156.34,
  rpsStatus: 'down_worse',
  errorDistribution: {
    'rpc error: code = Internal desc = Internal error.': 5,
    'rpc error: code = PermissionDenied desc = Permission denied.': 4
  },
  statusCodeDistribution: {
    'OK': 191
  },
  status: 'OK',
  'options': {
    'call': 'helloworld.Greeter.SayHello',
    'proto': '../../testdata/greeter.proto',
    'host': '0.0.0.0:50051',
    'n': 200,
    'c': 50,
    'timeout': 20000000000,
    'dialTimeout': 10000000000,
    'data': {
      'name': 'Joe'
    },
    'binary': false,
    'insecure': true,
    'CPUs': 8
  }
}, {
  id: 24,
  date: date4.toISOString(),
  projectId: 12,
  count: 400,
  total: 345959832,
  average: 43277742,
  averageStatus: 'up_better',
  fastest: 54304280,
  fastestStatus: 'down_better',
  slowest: 66684994,
  slowestStatus: 'up_better',
  rps: 2256.34,
  rpsStatus: 'down_worse',
  errorDistribution: {
    'rpc error: code = Internal desc = Internal error.': 5,
    'rpc error: code = PermissionDenied desc = Permission denied.': 4
  },
  statusCodeDistribution: {
    'OK': 191
  },
  status: 'OK',
  'options': {
    'call': 'helloworld.Greeter.SayHello',
    'proto': '../../testdata/greeter.proto',
    'host': '0.0.0.0:50051',
    'n': 200,
    'c': 50,
    'timeout': 20000000000,
    'dialTimeout': 10000000000,
    'data': {
      'name': 'Joe'
    },
    'binary': false,
    'insecure': true,
    'CPUs': 8
  }
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

async function getReport (id) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      const p = _.find(reports, p => p.id.toString() === id.toString())
      resolve(p)
    }, getRandomInt(800))
  })
}

export default class ReportContainer extends Container {
  constructor (props) {
    super(props)
    this.state = {
      reports: [],
      currentReport: {},
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

  async fetchReport (id) {
    this.setState(state => {
      return {
        ...state, isFetching: true
      }
    })

    try {
      const r = await getReport(id)
      this.setState({
        currentReport: r,
        isFetching: false
      })
    } catch (err) {
      console.log('error: ', err)
    }
  }
}
