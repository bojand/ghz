import React, { Component } from 'react'
import { Table, Heading, IconButton, Pane, Icon, Badge, Button } from 'evergreen-ui'
import { Link as RouterLink } from 'react-router-dom'
import { filter } from 'fuzzaldrin-plus'

import {
  Order,
  getIconForOrder
} from '../lib/common'

import EditProjectDialog from './EditProjectDialog'
import StatusBadge from './StatusBadge'

export default class ProjectList extends Component {
  constructor (props) {
    super(props)

    this.state = {
      searchQuery: '',
      ordering: Order.DESC,
      page: 0,
      editProjectVisible: false,
      editProject: null
    }
  }

  componentDidMount () {
    this.props.projectStore.fetchProjects()
  }

  componentDidUpdate (prevProps) {
    if (!this.props.projectStore.state.isFetching) {
      const currentList = this.props.projectStore.state.projects
      const prevList = prevProps.projectStore.state.projects

      if (currentList.length === 0 && prevList.length > 0) {
        this.setState({ page: 0 })
        this.props.projectStore.fetchProjects(this.state.ordering, 0)
      }
    }
  }

  sort () {
    const order = this.state.ordering === Order.ASC ? Order.DESC : Order.ASC
    this.props.projectStore.fetchProjects(order, this.state.page)
    this.setState({ ordering: order })
  }

  fetchPage (page) {
    if (page < 0) {
      page = 0
    }

    this.setState({ page })

    this.props.projectStore.fetchProjects(this.state.ordering, page)
  }

  // Filter the profiles based on the name property.
  filter (projects) {
    const searchQuery = this.state.searchQuery.trim()

    // If the searchQuery is empty, return the profiles as is.
    if (searchQuery.length === 0) return projects

    return projects.filter(p => {
      // Use the filter from fuzzaldrin-plus to filter by name.
      const result = filter([p.name], searchQuery)
      return result.length === 1
    })
  }

  handleFilterChange (value) {
    this.setState({ searchQuery: value })
  }

  handleEditProject (project) {
    this.setState({
      editProject: project,
      editProjectVisible: true
    })
  }

  render () {
    const { state: { projects, totalProjects } } = this.props.projectStore

    const items = this.filter(projects)

    const totalPerPage = 20

    return (
      <Pane>
        <Pane display='flex' alignItems='center'>
          <Heading size={600}>PROJECTS</Heading>
          {this.state.editProjectVisible
            ? <EditProjectDialog
              projectStore={this.props.projectStore}
              project={this.state.editProject}
              isShown={this.state.editProjectVisible}
              onDone={() => {
                this.setState({ editProjectVisible: false })
              }}
            /> : null
          }
          <Button marginLeft={14} iconBefore='plus' appearance='minimal' intent='none'
            onClick={() => {
              this.setState({ editProjectVisible: !this.state.editProjectVisible, editProject: null })
            }}>NEW</Button>
        </Pane>

        <Table marginY={24}>
          <Table.Head>
            <Table.TextHeaderCell maxWidth={80} textProps={{ size: 400 }}>
              <Pane display='flex'>
                ID
                <IconButton
                  marginLeft={10}
                  icon={getIconForOrder(this.state.ordering)}
                  appearance='minimal'
                  height={20}
                  onClick={() => this.sort()}
                />
              </Pane>
            </Table.TextHeaderCell>
            <Table.SearchHeaderCell maxWidth={260}
              onChange={v => this.handleFilterChange(v)}
              value={this.state.searchQuery}
            />
            <Table.TextHeaderCell textProps={{ size: 400 }}>
              Description
            </Table.TextHeaderCell>
            <Table.TextHeaderCell maxWidth={80} textProps={{ size: 400 }}>
              Status
            </Table.TextHeaderCell>
            <Table.HeaderCell maxWidth={50}>
            </Table.HeaderCell>
          </Table.Head>
          <Table.Body>
            {items.map(p => (
              <Table.Row key={p.id}>
                <Table.TextCell maxWidth={80} isNumber>
                  {p.id}
                </Table.TextCell>
                <Table.TextCell maxWidth={260} textProps={{ size: 400 }}>
                  <RouterLink to={`/projects/${p.id}`}>
                    {p.name}
                  </RouterLink>
                </Table.TextCell>
                <Table.TextCell textProps={{ size: 400 }}>{p.description}</Table.TextCell>
                <Table.TextCell
                  maxWidth={80}
                  display='flex' textAlign='center' alignItems='center'>
                  <StatusBadge status={p.status} marginRight={8} />
                </Table.TextCell>
                <Table.Cell maxWidth={50}>
                  <IconButton icon='edit' height={24} appearance='minimal' onClick={ev => this.handleEditProject(p)} />
                </Table.Cell>
              </Table.Row>
            ))}
          </Table.Body>
          <Pane justifyContent='right' marginTop={10} display='flex'>
            <IconButton
              disabled={totalProjects < totalPerPage || this.state.page === 0}
              icon='chevron-left'
              onClick={() => this.fetchPage(this.state.page - 1)}
            />
            <IconButton
              disabled={totalProjects < totalPerPage || projects.length < totalPerPage}
              marginLeft={10}
              icon='chevron-right'
              onClick={() => this.fetchPage(this.state.page + 1)}
            />
          </Pane>
        </Table>
      </Pane>
    )
  }
}
