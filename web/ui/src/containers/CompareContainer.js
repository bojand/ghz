import { Container } from 'unstated'
import ky from 'ky'
import { toaster } from 'evergreen-ui'

import { getAppRoot } from '../lib/common'

const api = ky.extend({ prefixUrl: getAppRoot() + '/api/reports/' })

export default class CompareContainer extends Container {
  constructor (props) {
    super(props)
    this.state = {
      report1: null,
      report2: null,
      isFetching: false
    }
  }

  async fetchReports (reportId1, reportId2) {
    this.setState({
      isFetching: true
    })

    let report1 = null
    let report2 = null

    try {
      report1 = await api.get(`${reportId1}`).json()
    } catch (err) {
      toaster.danger(err.message)
      console.log('error: ', err)
      return
    }

    try {
      if (reportId2.toLowerCase() === 'previous') {
        report2 = await api.get(`${reportId1}/previous`).json()
      } else {
        report2 = await api.get(`${reportId2}`).json()
      }
    } catch (err) {
      console.log('error: ', err)
    }

    this.setState({
      report1,
      report2,
      isFetching: false
    })
  }
}
