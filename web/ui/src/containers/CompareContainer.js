import { Container } from 'unstated'
import ky from 'ky'
import { toaster } from 'evergreen-ui'

const api = ky.extend({ prefixUrl: 'http://localhost:3000/api/reports/' })

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

    try {
      const report1 = await api.get(`${reportId1}`).json()

      let report2 = null
      if (reportId2.toLowerCase() === 'previous') {
        report2 = await api.get(`${reportId1}/previous`).json()
      } else {
        report2 = await api.get(`${reportId2}`).json()
      }

      this.setState({
        report1,
        report2,
        isFetching: false
      })
    } catch (err) {
      toaster.danger(err.message)
      console.log('error: ', err)
    }
  }
}
