import React from "react";
import PropTypes from "prop-types";
import Flippy, { FrontSide, BackSide } from "react-flippy-material-ui";
import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import Typography from "@material-ui/core/Typography";
import { Link as RouterLink } from "react-router-dom";
import Link from "@material-ui/core/Link";
import MiniChart from "../tradingview/MiniChart";

// TODO: figure out how to use props in stateless func - can we skip prop type?
// eslint-disable-next-line react/prefer-stateless-function
export default class Hit extends React.Component {
  render() {
    const { hit } = this.props;

    return (
      <div key={hit.symbol}>
        <Card>
          <CardContent>
            <Typography variant="h5" component="h2">
              {hit.symbol} ({hit.quoteDate}){" "}
              <Link component={RouterLink} to={`/symbol/${hit.symbol}`}>
                (detail)
              </Link>
            </Typography>
            <Flippy
              flipOnClick
              flipDirection="horizontal"
              style={{ height: "340px", width: "100%" }}
            >
              <FrontSide>
                <div style={{ float: "left", margin: "10px" }}>
                  <Typography variant="h6" component="h3">
                    {hit.companyName}
                  </Typography>
                  <Typography>
                    {hit.open} - {hit.close} ({hit.open / hit.close / hit.open})
                  </Typography>
                  <Typography>Volume: {hit.volume}</Typography>
                  <Typography>Market Cap: {hit.marketCap}</Typography>
                  <Typography>PE: {hit.pe}</Typography>
                  <Typography>Beta: {hit.beta}</Typography>
                  <Typography>
                    52 week h / l: {hit.high52Wk} / {hit.low52Wk}
                  </Typography>
                  <Typography>50ma: {hit.ma50}</Typography>
                  <Typography>200ma: {hit.ma200}</Typography>
                  <Typography>
                    Avg Vol 10 / 30: {hit.avg10Vol} / {hit.avg30Vol}
                  </Typography>
                </div>
              </FrontSide>
              <BackSide style={{ top: "0px" }}>
                <div style={{ float: "left", margin: "10px" }}>
                  <MiniChart symbol={hit.symbol} key={hit.symbol} />
                </div>
              </BackSide>
            </Flippy>
          </CardContent>
        </Card>
      </div>
    );
  }
}

Hit.propTypes = PropTypes.any.isRequired;
