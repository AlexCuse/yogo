import React from "react";
import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import Typography from "@material-ui/core/Typography";
import { Link as RouterLink } from "react-router-dom";
import Link from "@material-ui/core/Link";
import PropTypes from "prop-types";

export default function Summary({ signal }) {
  const sigpath = `/signals/${signal.name}/detail`;
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
        <Link component={RouterLink} to={sigpath}>
          Details
        </Link>
      </CardContent>
    </Card>
  );
}

Summary.propTypes = {
  signal: PropTypes.shape({
    name: PropTypes.string.isRequired,
    source: PropTypes.string.isRequired,
    count: PropTypes.number.isRequired,
  }).isRequired,
};
