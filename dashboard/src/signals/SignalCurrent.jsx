import React from "react";
import Signal from "./Signal";

export default class extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      signals: [],
    };

    this.watchApiUrl = new URL(
      "api/signal/currentbyname?name={}",
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
          <Signal signal={v} key={v.name} />
        ))}
      </div>
    );
  }
}
