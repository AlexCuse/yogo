import React from "react";
import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import Typography from "@material-ui/core/Typography";
import { Link as RouterLink } from "react-router-dom";
import Link from "@material-ui/core/Link";
import PropTypes from "prop-types";
import EdiText from "react-editext";

export default function Summary({ signal }) {
  const onSave = (val) => {
    const signalUrl = new URL(
      `api/signal`,
      process.env.REACT_APP_SIGNAL_API_URL
    );

    const s = {
      name: signal.name,
      source: val,
    };

    const requestOptions = {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(s),
    };

    fetch(signalUrl, requestOptions).then((response) => {
      // eslint-disable-next-line
      console.log("Saved Value -> ", requestOptions.body, response);
    });
  };

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
        <EdiText
          type="textarea"
          value={signal.source}
          inputProps={{ rows: 5 }}
          onSave={onSave}
        />
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
