import React, { Component } from 'react'
import { Table, Heading, IconButton, Pane, Icon, Button } from 'evergreen-ui'
import { Link as RouterLink } from 'react-router-dom'
import { filter } from 'fuzzaldrin-plus'

import {
  Order,
  getIconForOrder,
  getIconForStatus,
  getColorForStatus
} from '../lib/common'

import EditProjectDialog from './EditProjectDialog'

export default class ProjectList extends Component {
  constructor (props) {
    super(props)

    this.state = {
      searchQuery: '',
      ordering: Order.NONE,
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
        this.props.projectStore.fetchProjects()
      }
    }
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
    const { state: { projects } } = this.props.projectStore

    const items = this.filter(projects)

    return (
      <Pane>
        <Pane display='flex' alignItems='center'>
          <Heading size={500}>PROJECTS</Heading>
          {this.state.editProjectVisible
            ? <EditProjectDialog
              projectStore={this.props.projectStore}
              project={this.state.editProject}
              isShown={this.state.editProjectVisible}
              onDone={() => this.setState({ editProjectVisible: false })}
            /> : null
          }
          <Button onClick={() => this.setState({ editProjectVisible: !this.state.editProjectVisible })} marginLeft={14} iconBefore='plus' appearance='minimal' intent='none'>NEW</Button>
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
            <Table.TextHeaderCell maxWidth={100} textProps={{ size: 400 }}>
              Reports
            </Table.TextHeaderCell>
            <Table.TextHeaderCell maxWidth={80} textProps={{ size: 400 }}>
              Status
            </Table.TextHeaderCell>
            <Table.HeaderCell maxWidth={40}>
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
                <Table.TextCell maxWidth={100} isNumber>
                  {p.reports}
                </Table.TextCell>
                <Table.TextCell
                  maxWidth={80}
                  display='flex' textAlign='center'>
                  <Icon
                    icon={getIconForStatus(p.status)}
                    color={getColorForStatus(p.status)} />
                </Table.TextCell>
                <Table.Cell maxWidth={40}>
                  <IconButton icon='edit' height={24} appearance='minimal' onClick={ev => this.handleEditProject(p)} />
                </Table.Cell>
              </Table.Row>
            ))}
          </Table.Body>
        </Table>
      </Pane>
    )
  }
}
