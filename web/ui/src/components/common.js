const Order = {
  NONE: 'NONE',
  ASC: 'ASC',
  DESC: 'DESC'
}

function getIconForOrder (order) {
  switch (order) {
    case Order.ASC:
      return 'arrow-up'
    case Order.DESC:
      return 'arrow-down'
    default:
      return 'arrow-down'
  }
}

function getIconForStatus (status) {
  switch (status) {
    case 'OK':
      return 'tick-circle'
    default:
      return 'error'
  }
}

function getColorForStatus (status) {
  switch (status) {
    case 'OK':
      return 'success'
    default:
      return 'danger'
  }
}

function getIconForMetricStatus (status) {
  switch (status) {
    case 'up_better':
    case 'up_worse':
      return 'arrow-up'
    default:
      return 'arrow-down'
  }
}

module.exports = {
  getIconForMetricStatus,
  getIconForOrder,
  getIconForStatus,
  getColorForStatus,
  Order
}
