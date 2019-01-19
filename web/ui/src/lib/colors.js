import { getRandomInt } from '../lib/common'

const colors = {
  red: 'rgb(255, 0, 0)',
  orange: 'rgb(255, 128, 0)',
  yellow: 'rgb(255, 255, 0)',
  green: 'rgb(0, 151, 0)',
  blue: 'rgb(0, 128, 255)',
  purple: 'rgb(127, 0, 255)',
  grey: 'rgb(201, 203, 207)',
  teal: 'rgb(75, 192, 192)',
  // red: 'rgb(255, 99, 132)',
  darkYellow: 'rgb(255, 205, 86)',
  skyBlue: 'rgb(54, 162, 235)',
  lightPurple: 'rgb(153, 102, 255)',

}

function randomColor () {
  const r = getRandomInt(255)
  const g = getRandomInt(255)
  const b = getRandomInt(255)

  return `rgb(${r}, ${g}, ${b})`
}

module.exports = {
  colors,
  randomColor
}
