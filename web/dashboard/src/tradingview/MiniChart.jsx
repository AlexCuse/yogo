import React from "react";
import PropTypes from "prop-types";

// TODO: figure out how to use props in stateless func - can we skip prop type?
// eslint-disable-next-line react/prefer-stateless-function
export default class MiniChart extends React.Component {
  constructor(props) {
    super(props);
    this.ref = React.createRef();
  }

  componentDidMount() {
    const { symbol } = this.props;

    const script = document.createElement("script");
    script.src =
      "https://s3.tradingview.com/external-embedding/embed-widget-mini-symbol-overview.js";
    script.async = true;
    script.innerHTML = JSON.stringify({
      symbol,
      width: 500,
      height: 320,
      locale: "en",
      dateRange: "3M",
      colorTheme: "light",
      trendLineColor: "#37a6ef",
      underLineColor: "#E3F2FD",
      isTransparent: false,
      // autosize: false,
      largeChartUrl: "",
    });
    this.ref.current.appendChild(script);
  }

  render() {
    const { symbol } = this.props;

    return (
      <div key={symbol} className="tradingview-widget-container" ref={this.ref}>
        <div className="tradingview-widget-container__widget" />
        <div className="tradingview-widget-copyright">
          <a
            href={`https://www.tradingview.com/symbols/${symbol}`}
            rel="noreferrer"
            target="_blank"
          >
            <span className="blue-text">{symbol}</span>
          </a>{" "}
          from TradingView
        </div>
      </div>
    );
  }
}

MiniChart.propTypes = {
  symbol: PropTypes.string.isRequired,
};
