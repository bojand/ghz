import React, { Component } from 'react'
import { Table, Heading, Link, IconButton, Pane, Text } from 'evergreen-ui'
import { filter } from 'fuzzaldrin-plus'

import EditProjectDialog from './EditProjectDialog'

const Order = {
  NONE: 'NONE',
  ASC: 'ASC',
  DESC: 'DESC'
}

export default class ProjectList extends Component {
  constructor (props) {
    super(props)

    this.state = {
      searchQuery: '',
      orderedColumn: 1,
      ordering: Order.NONE,
      column2Show: 'email',
      projects: []
    }
  }

  componentDidMount () {
    this.props.projectStore.fetchProjects()
  }

  sort () {
    this.props.projectStore.fetchProjects(true)
    const order = this.state.ordering === Order.ASC ? Order.DESC : Order.ASC
    this.setState({ ordering: order })
  }

  // Filter the profiles based on the name property.
  filter (projects) {
    const searchQuery = this.state.searchQuery.trim()

    // If the searchQuery is empty, return the profiles as is.
    if (searchQuery.length === 0) return projects

    return projects.filter(profile => {
      // Use the filter from fuzzaldrin-plus to filter by name.
      const result = filter([profile.name], searchQuery)
      return result.length === 1
    })
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

  handleFilterChange (value) {
    this.setState({ searchQuery: value })
  }

  render () {
    const { state: { projects } } = this.props.projectStore

    return (
      <Pane>
        <Pane display='flex' alignItems='center'>
          <Heading size={500}>PROJECTS</Heading>
          <EditProjectDialog projectStore={this.props.projectStore} />
        </Pane>
        <Table marginY={24}>
          <Table.Head>
            <Table.TextHeaderCell maxWidth={80}>
              <Pane display='flex'>
                <Text>
                  ID
                </Text>
                <IconButton
                  marginLeft={10}
                  icon={this.getIconForOrder(this.state.ordering)}
                  appearance='minimal'
                  height={20}
                  onClick={() => this.sort()}
                />
              </Pane>
            </Table.TextHeaderCell>
            <Table.SearchHeaderCell maxWidth={250} />
            <Table.TextHeaderCell textProps={{ size: 400 }}>
              Description
            </Table.TextHeaderCell>
            <Table.TextHeaderCell maxWidth={100} textProps={{ size: 400 }}>
              Reports
            </Table.TextHeaderCell>
          </Table.Head>
          <Table.Body>
            {projects.map(p => (
              <Table.Row key={p.id} isSelectable onSelect={() => alert(p.name)}>
                <Table.TextCell maxWidth={80} isNumber>
                  {p.id}
                </Table.TextCell>
                <Table.TextCell maxWidth={250} textProps={{ size: 400 }}>
                  <Link href='#'>{p.name}</Link>
                </Table.TextCell>
                <Table.TextCell textProps={{ size: 400 }}>{p.description}</Table.TextCell>
                <Table.TextCell maxWidth={100} isNumber>
                  {p.reports}
                </Table.TextCell>
              </Table.Row>
            ))}
          </Table.Body>
        </Table>
      </Pane>
    )
  }
}
