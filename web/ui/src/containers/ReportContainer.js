import { Container } from 'unstated'
import ky from 'ky'
import { toaster } from 'evergreen-ui'

import { getAppRoot } from '../lib/common'

const api = ky.extend({ prefixUrl: getAppRoot() + '/api/' })

export default class ReportContainer extends Container {
  constructor (props) {
    super(props)
    this.state = {
      total: 0,
      reports: [],
      currentReport: {},
      isFetching: false
    }
  }

  async fetchReports (order = 'desc', sort = 'date', page = 0, projectId = 0) {
    this.setState({
      isFetching: true
    })

    const searchParams = new URLSearchParams()

    if (order) {
      searchParams.append('order', order)
      searchParams.append('sort', sort)
      searchParams.append('page', page)
    }

    try {
      let data
      let total
      if (!projectId) {
        const res = await api.get('reports', { searchParams }).json()
        data = res.data
        total = res.total
      } else {
        const res = await api.get(`projects/${projectId}/reports`, { searchParams }).json()
        data = res.data
        total = res.total
      }

      this.setState({
        total,
        reports: data,
        isFetching: false
      })
    } catch (err) {
      toaster.danger(err.message)
      console.log('error: ', err)
    }
  }

  async fetchReport (id) {
    this.setState({
      isFetching: true
    })

    try {
      const r = await api.get(`reports/${id}`).json()
      this.setState({
        currentReport: r,
        isFetching: false
      })
    } catch (err) {
      toaster.danger(err.message)
      console.log('error: ', err)
    }
  }

  async deleteReport (id) {
    this.setState({
      isFetching: true
    })

    try {
      await api.delete(`reports/${id}`).json()
      this.setState({
        isFetching: false
      })

      return true
    } catch (err) {
      toaster.danger(err.message)
      console.log('error: ', err)
    }
  }

  async deleteReports (ids) {
    this.setState({
      isFetching: true
    })

    try {
      const res = await api.post('reports/bulk_delete', { json: { ids } }).json()
      this.setState({
        isFetching: false
      })

      return res
    } catch (err) {
      toaster.danger(err.message)
      console.log('error: ', err)
    }
  }
}
