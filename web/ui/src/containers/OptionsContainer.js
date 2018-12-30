import { Container } from 'unstated'
import ky from 'ky'
import { toaster } from 'evergreen-ui'

const api = ky.extend({ prefixUrl: 'http://localhost:3000/api/reports/' })

export default class OptionsContainer extends Container {
  constructor (props) {
    super(props)
    this.state = {
      options: 0,
      isFetching: false
    }
  }

  async fetchOptions (reportId) {
    this.setState({
      isFetching: true
    })

    try {
      const { info } = await api.get(`${reportId}/options`).json()

      this.setState({
        options: info,
        isFetching: false
      })
    } catch (err) {
      toaster.danger(err.message)
      console.log('error: ', err)
    }
  }
}
