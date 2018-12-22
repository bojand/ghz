import React, { Component } from 'react'
import { Table, Heading, Link, IconButton, Pane, Icon } from 'evergreen-ui'

const Order = {
  NONE: 'NONE',
  ASC: 'ASC',
  DESC: 'DESC'
}

function formatNano (val) {
  return Number.parseFloat(val / 1000000).toFixed(2)
}

export default class ReportList extends Component {
  constructor (props) {
    super(props)

    this.state = {
      ordering: Order.NONE
    }
  }

  componentDidMount () {
    this.props.reportStore.fetchReports()
  }

  sort () {
    this.props.reportStore.fetchReports(true)
    const order = this.state.ordering === Order.ASC ? Order.DESC : Order.ASC
    this.setState({ ordering: order })
  }

  getIconForOrder (order) {
    switch (order) {
      case Order.ASC:
        return 'arrow-up'
      case Order.DESC:
        return 'arrow-down'
      default:
        return 'arrow-down'
    }
  }

  getIconForStatus (status) {
    switch (status) {
      case 'OK':
        return 'tick-circle'
      default:
        return 'error'
    }
  }

  getIconForMetricStatus (status) {
    switch (status) {
      case 'up_better':
      case 'up_worse':
        return 'arrow-up'
      default:
        return 'arrow-down'
    }
  }

  getColorForStatus (status) {
    switch (status) {
      case 'OK':
      case 'up_better':
      case 'down_better':
        return 'success'
      default:
        return 'danger'
    }
  }

  render () {
    const { state: { reports } } = this.props.reportStore

    return (
      <Pane {...this.props}>
        <Pane display='flex' alignItems='center' marginTop={6}>
          <Heading size={500}>REPORTS</Heading>
        </Pane>

        <Table marginY={30}>
          <Table.Head>
            <Table.TextHeaderCell minWidth={210} textProps={{ size: 400 }}>
              <Pane display='flex'>
                Date
                <IconButton
                  marginLeft={10}
                  icon={this.getIconForOrder(this.state.ordering)}
                  appearance='minimal'
                  height={20}
                  onClick={() => this.sort()}
                />
              </Pane>
            </Table.TextHeaderCell>
            <Table.TextHeaderCell textProps={{ size: 400 }}>
              Count
            </Table.TextHeaderCell>
            <Table.TextHeaderCell textProps={{ size: 400 }}>
              Total
            </Table.TextHeaderCell>
            <Table.TextHeaderCell textProps={{ size: 400 }}>
              Average
            </Table.TextHeaderCell>
            <Table.TextHeaderCell textProps={{ size: 400 }}>
              Slowest
            </Table.TextHeaderCell>
            <Table.TextHeaderCell textProps={{ size: 400 }}>
              Fastest
            </Table.TextHeaderCell>
            <Table.TextHeaderCell textProps={{ size: 400 }}>
              RPS
            </Table.TextHeaderCell>
            <Table.TextHeaderCell maxWidth={80}>
              Status
            </Table.TextHeaderCell>
          </Table.Head>
          <Table.Body>
            {reports.map(p => (
              <Table.Row key={p.id}>
                <Table.TextCell minWidth={210} >
                  <Link href='#'>{p.date}</Link>
                </Table.TextCell>
                <Table.TextCell isNumber>
                  {p.count}
                </Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNano(p.total)} ms
                </Table.TextCell>
                <Table.TextCell isNumber>
                  <Pane display='flex'>
                    {formatNano(p.average)} ms
                    <Icon
                      marginLeft={10}
                      icon={this.getIconForMetricStatus(p.averageStatus)}
                      color={this.getColorForStatus(p.averageStatus)} />
                  </Pane>
                </Table.TextCell>
                <Table.TextCell isNumber>
                  <Pane display='flex'>
                    {formatNano(p.slowest)} ms
                    <Icon
                      marginLeft={10}
                      icon={this.getIconForMetricStatus(p.slowestStatus)}
                      color={this.getColorForStatus(p.slowestStatus)} />
                  </Pane>
                </Table.TextCell>
                <Table.TextCell isNumber>
                  <Pane display='flex'>
                    {formatNano(p.fastest)} ms
                    <Icon
                      marginLeft={10}
                      icon={this.getIconForMetricStatus(p.fastestStatus)}
                      color={this.getColorForStatus(p.fastestStatus)} />
                  </Pane>
                </Table.TextCell>
                <Table.TextCell isNumber>
                  <Pane display='flex'>
                    {p.rps}
                    <Icon
                      marginLeft={10}
                      icon={this.getIconForMetricStatus(p.rpsStatus)}
                      color={this.getColorForStatus(p.rpsStatus)} />
                  </Pane>
                </Table.TextCell>
                <Table.TextCell
                  display='flex' textAlign='center' maxWidth={80}>
                  <Icon
                    icon={this.getIconForStatus(p.status)}
                    color={this.getColorForStatus(p.status)} />
                </Table.TextCell>
              </Table.Row>
            ))}
          </Table.Body>
        </Table>
      </Pane>
    )
  }
}
