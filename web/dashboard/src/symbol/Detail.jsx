import React from "react";
import PropTypes from "prop-types";
import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import Typography from "@material-ui/core/Typography";
import Chart from "../tradingview/Chart";

// TODO: figure out how to use props in stateless func - can we skip prop type?
// eslint-disable-next-line react/prefer-stateless-function
export default class Detail extends React.Component {
  constructor(props) {
    super(props);

    const { symbol } = props.match.params;

    this.state = { symbol };
    /*
    this.signalUrl = new URL(
      `api/signal/${name}/current`,
      process.env.REACT_APP_SIGNAL_API_URL
    );

    this.fetchSignal();
    */
  }

  render() {
    const { symbol } = this.state;

    return (
      <Card key={symbol}>
        <CardContent>
          <Typography variant="h4">{symbol}</Typography>
          <div>
            <Chart symbol={symbol} key={symbol} />
          </div>
        </CardContent>
      </Card>
    );
  }
}

Detail.propTypes = PropTypes.any.isRequired;
