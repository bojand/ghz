/**
 * Copyright (c) 2017-present, Facebook, Inc.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

const React = require('react');

const CompLibrary = require('../../core/CompLibrary.js');

const MarkdownBlock = CompLibrary.MarkdownBlock; /* Used to read markdown */
const Container = CompLibrary.Container;
const GridBlock = CompLibrary.GridBlock;

const siteConfig = require(process.cwd() + '/siteConfig.js');

function imgUrl(img) {
  return siteConfig.baseUrl + 'img/' + img;
}

class HomeSplash extends React.Component {
  render() {
    const { siteConfig, language = '' } = this.props;
    const { baseUrl, docsUrl } = siteConfig;
    const docsPart = `${docsUrl ? `${docsUrl}/` : ''}`;
    const langPart = `${language ? `${language}/` : ''}`;
    const docUrl = doc => `${baseUrl}${docsPart}${langPart}${doc}`;

    const SplashContainer = props => (
      <div className="homeContainer">
        <div className="homeSplashFade">
          <div className="wrapper homeWrapper">{props.children}</div>
        </div>
      </div>
    );

    const Logo = props => (
      <div>
        <br />
        <img src={props.img_src} alt="Logo" width="100" />
      </div>
    );

    const ProjectTitle = () => (
      <h2 className="projectTitle">
        <small>{siteConfig.tagline}</small>
      </h2>
    );

    const PromoSection = props => (
      <div className="section promoSection">
        <div className="promoRow">
          <div className="pluginRowBlock">{props.children}</div>
        </div>
      </div>
    );

    const Button = props => (
      <div className="pluginWrapper buttonWrapper">
        <a className="button" href={props.href} target={props.target}>
          {props.children}
        </a>
      </div>
    );

    return (
      <SplashContainer>
        <Logo img_src={`${baseUrl}img/green_fwd2.png`} />
        <div className="inner">
          <ProjectTitle siteConfig={siteConfig} />
          <PromoSection>
            <Button href={docUrl('intro.html')}>Get Started</Button>
            <Button href={`${siteConfig.repoUrl}`}>GitHub</Button>
          </PromoSection>
        </div>
      </SplashContainer>
    );
  }
}

class Index extends React.Component {
  render() {
    const { config: siteConfig, language = '' } = this.props;
    const { baseUrl } = siteConfig;

    const Block = props => (
      <Container
        padding={[ 'top' ]}
        id={props.id}
        background={props.background}>
        <GridBlock
          align="center"
          contents={props.children}
          layout={props.layout}
        />
      </Container>
    );

    const Features = () => (
      <Block layout="threeColumn">
        {[
          {
            content: 'Use proto file, or prebuilt protoset bundle, or server reflection',
            imageAlign: 'top',
            title: 'Use Proto, Protoset or Reflection',
          },
          {
            content: 'View test results in various fomats including CLI, CSV, JSON, HTML and InfluxData',
            imageAlign: 'top',
            title: 'Various Report Formats',
          },
          {
            content: 'Add custom data to requests using standard Go template variables',
            imageAlign: 'top',
            title: 'Custom Data',
          },
          {
            content: 'Test unary, streaming and duplex calls <br />using JSON or binary data',
            imageAlign: 'top',
            title: 'Flexible and featureful',
          },
          {
            content: 'Save, track, view and analyse test results <br />in a complementary web application <br /><a href="https://ghz-demo.herokuapp.com">Demo</a>',
            imageAlign: 'top',
            title: 'Complementary Web Application (Beta)',
          }
        ]}
      </Block>
    );

    const Badges = () => (
      <div className="productShowcaseSection" style={{ textAlign: 'center' }}>
          <a href={"https://github.com/bojand/ghz/releases/latest"}>
            <img src={"https://img.shields.io/github/release/bojand/ghz.svg?style=flat-square"} alt={"Release"} />
          </a>
          <a href={"https://github.com/bojand/ghz/actions?workflow=build"} style={{ margin: '5px' }}>
            <img src={"https://github.com/bojand/ghz/workflows/build/badge.svg"} alt={"Build status"} />
          </a>
          <a href={"https://goreportcard.com/report/github.com/bojand/ghz"}>
            <img src={"https://goreportcard.com/badge/github.com/bojand/ghz?style=flat-square"} alt={"Go Report Card"} />
          </a>
          <a href={"https://raw.githubusercontent.com/bojand/ghz/master/LICENSE"} style={{ margin: '5px' }}>
            <img src={"https://img.shields.io/github/license/bojand/ghz.svg?style=flat-square"} alt={"License"} />
          </a>
          <a href={"https://www.paypal.me/bojandj"}>
            <img src={"https://img.shields.io/badge/Donate-PayPal-green.svg?style=flat-square"} alt={"Donate"} />
          </a>
          <a href={"https://www.buymeacoffee.com/bojand"} style={{ margin: '5px' }}>
            <img src={"https://img.shields.io/badge/buy%20me-a%20coffee-orange.svg?style=flat-square"} alt={"Buy me a coffee"} />
          </a>
      </div>
    );

    const Description = () => (
      <Container>
        <div className="productShowcaseSection">
          <br />
          <img src={imgUrl('ghz_cobalt_plain.png')} alt="ghz" />
        </div>
      </Container>
    )

    const Screen = (props) => (
      <Container background={props.background}>
        <div className="productShowcaseSection">
          <br />
          <a href={imgUrl(props.image)}>
            <img width="860" src={imgUrl(props.preview)} style={{ border: '1px solid #d6dbdf'}} />
          </a>
          <br />
          <br />
        </div>
      </Container>
    )

    return (
      <div>
        <HomeSplash siteConfig={siteConfig} language={language} />
        <div className="mainContainer">
          <Badges />
          <Features />
          <Description />
          <Screen background='light' preview='project_detail_page_preview.png' image='project_detail_page.png' />
          <Screen preview='report_detail_page_preview.png' image='report_detail_page.png' />
        </div>
      </div>
    );
  }
}

module.exports = Index;
