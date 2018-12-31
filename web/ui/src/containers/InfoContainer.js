import { Container } from 'unstated'
import ky from 'ky'
import { toaster } from 'evergreen-ui'

import { getAppRoot } from '../lib/common'

const api = ky.extend({ prefixUrl: getAppRoot() + '/api/' })

export default class InfoContainer extends Container {
  constructor (props) {
    super(props)

    this.state = {
      info: null,
      isFetching: false
    }
  }

  async fetchInfo () {
    this.setState({
      isFetching: true
    })

    try {
      const info = await api.get('info').json()

      this.setState({
        info,
        isFetching: false
      })
    } catch (err) {
      toaster.danger(err.message)
      console.log('error: ', err)
    }
  }
}
