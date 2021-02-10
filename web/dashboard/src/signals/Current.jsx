import React from "react";
import Summary from "./Summary";

export default class extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      signals: [],
    };

    this.watchApiUrl = new URL(
      "api/signals/current",
      process.env.REACT_APP_SIGNAL_API_URL
    );

    this.fetchWatchList();
  }

  fetchWatchList() {
    fetch(this.watchApiUrl)
      .then((response) => response.json())
      .then((data) => {
        this.setState({
          signals: data.map((w) => w),
        });
      });
  }

  render() {
    const { signals } = this.state;

    return (
      <div>
        {signals.map((v) => (
          <Summary signal={v} key={v.name} />
        ))}
      </div>
    );
  }
}
