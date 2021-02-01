import React from "react";
import Card from "@material-ui/core/Card";
import CardActions from "@material-ui/core/CardActions";
import CardContent from "@material-ui/core/CardContent";
import Button from "@material-ui/core/Button";
import Typography from "@material-ui/core/Typography";
import PropTypes from "prop-types";
import MiniChart from "../tradingview/MiniChart";

export default function Watch({ symbol }) {
  return (
    <Card>
      <CardContent>
        <Typography variant="h5" component="h2">
          {symbol}
        </Typography>
        <MiniChart symbol={symbol} />
      </CardContent>
      <CardActions>
        <Button size="small">Learn More</Button>
      </CardActions>
    </Card>
  );
}

Watch.propTypes = {
  symbol: PropTypes.string.isRequired,
};
