import React from "react";
import PropTypes from "prop-types";

export default class Chart extends React.Component {
  componentDidMount() {
    const { symbol } = this.props;

    // eslint-disable-next-line
    new TradingView.widget({
      height: 600,
      width: 750,
      symbol: `${symbol}`,
      interval: "D",
      timezone: "Etc/UTC",
      theme: "light",
      style: "1",
      locale: "en",
      toolbar_bg: "#f1f3f6",
      enable_publishing: false,
      allow_symbol_change: false,
      container_id: `tradingview_${symbol}`,
    });
  }

  render() {
    const { symbol } = this.props;

    return (
      <div className="tradingview-widget-container">
        <div id={`tradingview_${symbol}`} />
      </div>
    );
  }
}

Chart.propTypes = {
  symbol: PropTypes.string.isRequired,
};
