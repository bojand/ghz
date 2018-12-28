import { Container } from 'unstated'
import _ from 'lodash'

import { getRandomInt } from '../lib/common'

const date1 = new Date()
const date2 = new Date()
const date3 = new Date()
const date4 = new Date()

date2.setMonth(date1.getMonth() - 1)
date3.setMonth(date2.getMonth() - 1)
date4.setMonth(date3.getMonth() - 1)

const reports = [{
  name: 'staging-oss.eventapi.createEvent',
  id: 10,
  date: date1.toISOString(),
  projectId: 11,
  count: 200,
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
    'rpc error: code = Internal desc = Internal error.': 7,
    'rpc error: code = PermissionDenied desc = Permission denied.': 2,
    'rpc error: code = Unauthorized desc = Unauthorized.': 1,
    'rpc error: code = DeadlineExceeded desc = Deadline exceeded.': 3,
    'rpc error: code = ResourceExhausted desc = Resource exhausted.': 4
  },
  statusCodeDistribution: {
    'OK': 191,
    'Canceled': 7,
    'ResourceExhausted': 3,
    'FailedPrecondition': 3
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
  },
  'latencyDistribution': [
    {
      'percentage': 10,
      'latency': 28213186
    },
    {
      'percentage': 25,
      'latency': 29888900
    },
    {
      'percentage': 50,
      'latency': 31946130
    },
    {
      'percentage': 75,
      'latency': 34774814
    },
    {
      'percentage': 90,
      'latency': 35910250
    },
    {
      'percentage': 95,
      'latency': 57698757
    },
    {
      'percentage': 99,
      'latency': 62984994
    }
  ],
  'histogram': [
    {
      'mark': 0.02540428,
      'count': 1,
      'frequency': 0.005235602094240838
    },
    {
      'mark': 0.0291623514,
      'count': 34,
      'frequency': 0.17801047120418848
    },
    {
      'mark': 0.0329204228,
      'count': 79,
      'frequency': 0.41361256544502617
    },
    {
      'mark': 0.0366784942,
      'count': 61,
      'frequency': 0.3193717277486911
    },
    {
      'mark': 0.0404365656,
      'count': 6,
      'frequency': 0.031413612565445025
    },
    {
      'mark': 0.044194637,
      'count': 0,
      'frequency': 0
    },
    {
      'mark': 0.0479527084,
      'count': 0,
      'frequency': 0
    },
    {
      'mark': 0.0517107798,
      'count': 0,
      'frequency': 0
    },
    {
      'mark': 0.0554688512,
      'count': 1,
      'frequency': 0.005235602094240838
    },
    {
      'mark': 0.0592269226,
      'count': 5,
      'frequency': 0.02617801047120419
    },
    {
      'mark': 0.062984994,
      'count': 4,
      'frequency': 0.020942408376963352
    }
  ],
  tags: {
    env: 'staging',
    'created by': 'Joe Developer'
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
  },
  'latencyDistribution': [
    {
      'percentage': 10,
      'latency': 28213186
    },
    {
      'percentage': 25,
      'latency': 29888900
    },
    {
      'percentage': 50,
      'latency': 31946130
    },
    {
      'percentage': 75,
      'latency': 34774814
    },
    {
      'percentage': 90,
      'latency': 35910250
    },
    {
      'percentage': 95,
      'latency': 57698757
    },
    {
      'percentage': 99,
      'latency': 62984994
    }
  ],
  'histogram': [
    {
      'mark': 0.02540428,
      'count': 1,
      'frequency': 0.005235602094240838
    },
    {
      'mark': 0.0291623514,
      'count': 34,
      'frequency': 0.17801047120418848
    },
    {
      'mark': 0.0329204228,
      'count': 79,
      'frequency': 0.41361256544502617
    },
    {
      'mark': 0.0366784942,
      'count': 61,
      'frequency': 0.3193717277486911
    },
    {
      'mark': 0.0404365656,
      'count': 6,
      'frequency': 0.031413612565445025
    },
    {
      'mark': 0.044194637,
      'count': 0,
      'frequency': 0
    },
    {
      'mark': 0.0479527084,
      'count': 0,
      'frequency': 0
    },
    {
      'mark': 0.0517107798,
      'count': 0,
      'frequency': 0
    },
    {
      'mark': 0.0554688512,
      'count': 1,
      'frequency': 0.005235602094240838
    },
    {
      'mark': 0.0592269226,
      'count': 5,
      'frequency': 0.02617801047120419
    },
    {
      'mark': 0.062984994,
      'count': 4,
      'frequency': 0.020942408376963352
    }
  ]
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
  },
  'latencyDistribution': [
    {
      'percentage': 10,
      'latency': 28213186
    },
    {
      'percentage': 25,
      'latency': 29888900
    },
    {
      'percentage': 50,
      'latency': 31946130
    },
    {
      'percentage': 75,
      'latency': 34774814
    },
    {
      'percentage': 90,
      'latency': 35910250
    },
    {
      'percentage': 95,
      'latency': 57698757
    },
    {
      'percentage': 99,
      'latency': 62984994
    }
  ],
  'histogram': [
    {
      'mark': 0.02540428,
      'count': 1,
      'frequency': 0.005235602094240838
    },
    {
      'mark': 0.0291623514,
      'count': 34,
      'frequency': 0.17801047120418848
    },
    {
      'mark': 0.0329204228,
      'count': 79,
      'frequency': 0.41361256544502617
    },
    {
      'mark': 0.0366784942,
      'count': 61,
      'frequency': 0.3193717277486911
    },
    {
      'mark': 0.0404365656,
      'count': 6,
      'frequency': 0.031413612565445025
    },
    {
      'mark': 0.044194637,
      'count': 0,
      'frequency': 0
    },
    {
      'mark': 0.0479527084,
      'count': 0,
      'frequency': 0
    },
    {
      'mark': 0.0517107798,
      'count': 0,
      'frequency': 0
    },
    {
      'mark': 0.0554688512,
      'count': 1,
      'frequency': 0.005235602094240838
    },
    {
      'mark': 0.0592269226,
      'count': 5,
      'frequency': 0.02617801047120419
    },
    {
      'mark': 0.062984994,
      'count': 4,
      'frequency': 0.020942408376963352
    }
  ]
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
  },
  'latencyDistribution': [
    {
      'percentage': 10,
      'latency': 28213186
    },
    {
      'percentage': 25,
      'latency': 29888900
    },
    {
      'percentage': 50,
      'latency': 31946130
    },
    {
      'percentage': 75,
      'latency': 34774814
    },
    {
      'percentage': 90,
      'latency': 35910250
    },
    {
      'percentage': 95,
      'latency': 57698757
    },
    {
      'percentage': 99,
      'latency': 62984994
    }
  ],
  'histogram': [
    {
      'mark': 0.02540428,
      'count': 1,
      'frequency': 0.005235602094240838
    },
    {
      'mark': 0.0291623514,
      'count': 34,
      'frequency': 0.17801047120418848
    },
    {
      'mark': 0.0329204228,
      'count': 79,
      'frequency': 0.41361256544502617
    },
    {
      'mark': 0.0366784942,
      'count': 61,
      'frequency': 0.3193717277486911
    },
    {
      'mark': 0.0404365656,
      'count': 6,
      'frequency': 0.031413612565445025
    },
    {
      'mark': 0.044194637,
      'count': 0,
      'frequency': 0
    },
    {
      'mark': 0.0479527084,
      'count': 0,
      'frequency': 0
    },
    {
      'mark': 0.0517107798,
      'count': 0,
      'frequency': 0
    },
    {
      'mark': 0.0554688512,
      'count': 1,
      'frequency': 0.005235602094240838
    },
    {
      'mark': 0.0592269226,
      'count': 5,
      'frequency': 0.02617801047120419
    },
    {
      'mark': 0.062984994,
      'count': 4,
      'frequency': 0.020942408376963352
    }
  ]
}]

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
