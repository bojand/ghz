import { Container } from 'unstated'
import ky from 'ky'
import { toaster } from 'evergreen-ui'

const api = ky.extend({ prefixUrl: 'http://localhost:3000/api/reports/' })

export default class LatencyContainer extends Container {
  constructor (props) {
    super(props)
    this.state = {
      latencyDistribution: 0,
      isFetching: false
    }
  }

  async fetchLatencyDistribution (reportId) {
    this.setState({
      isFetching: true
    })

    try {
      const { list } = await api.get(`${reportId}/latencies`).json()

      this.setState({
        latencyDistribution: list,
        isFetching: false
      })
    } catch (err) {
      toaster.danger(err.message)
      console.log('error: ', err)
    }
  }
}
