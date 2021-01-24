import React from "react";
import PropTypes from "prop-types";

export default class Tradeview extends React.Component {
  componentDidMount() {
    const { symbol } = this.props;

    // eslint-disable-next-line
    new TradingView.widget({
      autosize: true,
      symbol: `${symbol}`,
      interval: "5",
      timezone: "Etc/UTC",
      theme: "light",
      style: "1",
      locale: "en",
      toolbar_bg: "#f1f3f6",
      enable_publishing: false,
      allow_symbol_change: true,
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

Tradeview.propTypes = {
  symbol: PropTypes.string.isRequired,
};
