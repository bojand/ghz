import { Container } from 'unstated'
import ky from 'ky'
import { toaster } from 'evergreen-ui'

import { getAppRoot } from '../lib/common'

const api = ky.extend({ prefixUrl: getAppRoot() + '/api/reports/' })

export default class HistogramContainer extends Container {
  constructor (props) {
    super(props)
    this.state = {
      histogram: 0,
      isFetching: false
    }
  }

  async fetchHistogram (reportId) {
    this.setState({
      isFetching: true
    })

    try {
      const { buckets } = await api.get(`${reportId}/histogram`).json()

      this.setState({
        histogram: buckets,
        isFetching: false
      })
    } catch (err) {
      toaster.danger(err.message)
      console.log('error: ', err)
    }
  }
}
