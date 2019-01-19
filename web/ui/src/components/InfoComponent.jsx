import React, { Component } from 'react'
import { Pane, Text, Strong, Heading, Link, Paragraph } from 'evergreen-ui'
import _ from 'lodash'

export default class InfoComponent extends Component {
  componentDidMount () {
    this.props.infoStore.fetchInfo()
  }

  toWord (v) {
    return _.chain(v).upperFirst().words().value().join(' ')
  }

  render () {
    const { state: { info } } = this.props.infoStore

    if (!info) {
      return (<Pane />)
    }

    let infoKey = 0

    return (
      <Pane marginTop={8}>

        <Heading size={600} marginBottom={4}>
          General Information
        </Heading>
        <Paragraph>
          <Link href='http://ghz.sh/' target='_blank'>
            Website
          </Link>
        </Paragraph>
        <Paragraph>
          <Link href='https://github.com/bojand/ghz' target='_blank'>
            Github
          </Link>
        </Paragraph>

        <Pane marginTop={16}>
          <Heading size={500} marginBottom={4}>
            Donate <small>❤️</small>
          </Heading>
          <Paragraph>
            <Link href='https://www.paypal.me/bojandj' target='_blank'>
              PayPal
            </Link>
          </Paragraph>
          <Paragraph>
            <Link href='https://www.buymeacoffee.com/bojand' target='_blank'>
              Buy Me A Coffee
            </Link>
          </Paragraph>
        </Pane>

        <Pane marginTop={16}>
          <Heading size={600} marginBottom={4}>
            Application
          </Heading>
          {_.map(info, (v, k) => (
            <Pane key={`infoKey-${infoKey++}`}>
              {_.isString(k) && !_.isObject(v) && v
                ? <Text>
                  <Strong>{this.toWord(k)}:</Strong> {v}
                </Text>
                : null}
            </Pane>
          ))}
        </Pane>

        <Pane marginTop={16}>
          <Heading size={600} marginBottom={4}>
            Memory Info
          </Heading>
          {_.map(info.memoryInfo, (v, k) => (
            <Pane key={`infoKey-${infoKey++}`}>
              {_.isString(k) && !_.isObject(v) && v
                ? <Text>
                  <Strong>{this.toWord(k)}:</Strong> {v}
                </Text>
                : null}
            </Pane>
          ))}
        </Pane>
      </Pane>
    )
  }
}
