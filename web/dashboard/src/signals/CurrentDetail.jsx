import React from "react";
import PropTypes from "prop-types";
import Summary from "./Summary";
import Hit from "./Hit";

export default class CurrentDetail extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      signal: { name: "unloaded", source: "", active: [] },
    };

    const { name } = props.match.params;

    this.signalUrl = new URL(
      `api/signal/${name}/current`,
      process.env.REACT_APP_SIGNAL_API_URL
    );

    this.fetchSignal();
  }

  fetchSignal() {
    fetch(this.signalUrl)
      .then((response) => response.json())
      .then((data) => {
        this.setState({
          signal: data,
        });
      });
  }

  render() {
    const { signal } = this.state;
    const sig = {
      name: signal.name,
      source: signal.source,
      count: signal.active.length,
    };
    return (
      <div>
        <div>
          <Summary signal={sig} key={sig.name} />
        </div>
        <div>
          {signal.active.map((v) => (
            <Hit hit={v} key={v.symbol} />
          ))}
        </div>
      </div>
    );
  }
}

CurrentDetail.propTypes = PropTypes.any.isRequired;

/*
CurrentDetail.propTypes = PropTypes.shape({
  match: PropTypes.shape({
    params: PropTypes.shape({
      name: PropTypes.string.isRequired,
    }).isRequired,
  }).isRequired,
});
*/
