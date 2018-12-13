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
      <div className="projectLogo">
        {/* <img src={props.img_src} alt="Project Logo" /> */}
        <img src={imgUrl('ghz_cobalt_plain.png')} alt="ghz" />
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
        {/* <Logo img_src={`${baseUrl}img/green_fwd2.svg`} /> */}
        <div className="inner">
          <ProjectTitle siteConfig={siteConfig} />
          <img src={imgUrl('ghz_cobalt_plain.png')} alt="ghz" />
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
        padding={[ 'bottom', 'top' ]}
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
      <Block layout="fourColumn">
        {[
          {
            content: 'Use proto files with import paths, or prebuilt protoset bundle files',
            // image: `${baseUrl}img/green_fwd2.svg`,
            imageAlign: 'top',
            title: 'Proto and Protoset',
          },
          {
            content: 'View test results in various fomats including CLI, CSV, JSON, HTML and InfluxData',
            // image: `${baseUrl}img/green_fwd2.svg`,
            imageAlign: 'top',
            title: 'Various Reports',
          },
          {
            content: 'Add custom data to requests using standard Go template variables',
            // image: `${baseUrl}img/green_fwd2.svg`,
            imageAlign: 'top',
            title: 'Custom Data',
          },
          {
            content: 'Test unary, streaming and duplex call types',
            // image: `${baseUrl}img/green_fwd2.svg`,
            imageAlign: 'top',
            title: 'Flexible and complete',
          }
        ]}
      </Block>
    );

    const Badges = () => (
      <div className="productShowcaseSection" style={{ textAlign: 'center' }}>
          <a href={"https://github.com/bojand/ghz/releases/latest"}>
            <img src={"https://img.shields.io/github/release/bojand/ghz.svg?style=flat-square"} alt={"Release"} />
          </a>
          <a href={"https://travis-ci.org/bojand/ghz"} style={{ margin: '5px' }}>
            <img src={"https://img.shields.io/travis/bojand/ghz/master.svg?style=flat-square"} alt={"build status"} />
          </a>
          <a href={"https://goreportcard.com/report/github.com/bojand/ghz"}>
            <img src={"https://goreportcard.com/badge/github.com/bojand/ghz?style=flat-square"} alt={"Go Report Card"} />
          </a>
          <a href={"https://raw.githubusercontent.com/bojand/ghz/master/LICENSE"}>
            <img src={"https://img.shields.io/github/license/bojand/ghz.svg?style=flat-square"} alt={"License"} />
          </a>
          <a href={"https://www.paypal.me/bojandj"} style={{ margin: '5px' }}>
            <img src={"https://img.shields.io/badge/Donate-PayPal-green.svg?style=flat-square"} alt={"Donate"} />
          </a>
          <a href={"https://www.buymeacoffee.com/bojand"}>
            <img src={"https://img.shields.io/badge/buy%20me-a%20coffee-orange.svg?style=flat-square"} alt={"Buy me a coffee"} />
          </a>
      </div>
    );

    return (
      <div>
        <HomeSplash siteConfig={siteConfig} language={language} />
        <div className="mainContainer">
          <Badges />
          <Features />
          {/* <FeatureCallout />
          <LearnHow />
          <TryOut />
          <Description />
          <Showcase /> */}
        </div>
      </div>
    );
  }
}

module.exports = Index;
