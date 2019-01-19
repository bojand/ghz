import React from 'react'
import { Badge } from 'evergreen-ui'

const StatusBadge = ({ status, ...props }) => {
  if (status && status.toLowerCase() === 'ok') {
    return (
      <Badge color='green' isSolid {...props}>OK</Badge>
    )
  }

  return (
    <Badge color='red' isSolid {...props}>fail</Badge>
  )
}

export default StatusBadge
