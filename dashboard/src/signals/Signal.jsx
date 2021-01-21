import React from "react";
import Card from "@material-ui/core/Card";
import CardActions from "@material-ui/core/CardActions";
import CardContent from "@material-ui/core/CardContent";
import Button from "@material-ui/core/Button";
import Typography from "@material-ui/core/Typography";
import PropTypes from "prop-types";

export default function Signal({ signal }) {
  return (
    <Card>
      <CardContent>
        <Typography variant="h5" component="h2">
          {signal.name}
        </Typography>
        <Typography>
          hits: <strong>{signal.count}</strong>
        </Typography>
        <Typography>
          source: <i>{signal.source}</i>
        </Typography>
      </CardContent>
      <CardActions>
        <Button size="small">Learn More</Button>
      </CardActions>
    </Card>
  );
}

Signal.propTypes = {
  signal: PropTypes.shape({ name: PropTypes.string.isRequired, source: PropTypes.string.isRequired, count: PropTypes.number.isRequired }).isRequired,
};
