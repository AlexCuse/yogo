import React from "react";
import PropTypes from "prop-types";
import Flippy, { FrontSide, BackSide } from "react-flippy-material-ui";
import Chart from "../tradingview/Chart";

// TODO: figure out how to use props in stateless func - can we skip prop type?
// eslint-disable-next-line react/prefer-stateless-function
export default class Hit extends React.Component {
  render() {
    const { symbol } = this.props;

    return (
      <div key={symbol}>
        <Flippy
          flipOnClick
          flipDirection="horizontal"
          style={{ height: "250px", width: "50%" }}
        >
          <FrontSide>{symbol}</FrontSide>
          <BackSide>
            <Chart symbol={symbol} key={symbol} />
          </BackSide>
        </Flippy>
      </div>
    );
  }
}

Hit.propTypes = {
  symbol: PropTypes.string.isRequired,
};
