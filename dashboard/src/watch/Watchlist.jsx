import React from "react";
import Watch from "./Watch";

export default class extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      symbols: [],
    };

    this.watchApiUrl = new URL(
      "api/watch",
      process.env.REACT_APP_WATCH_API_URL
    );

    this.fetchWatchList();
  }

  fetchWatchList() {
    fetch(this.watchApiUrl)
      .then((response) => response.json())
      .then((data) => {
        this.setState({
          symbols: data.map((w) => w.symbol),
        });
      });
  }

  render() {
    const { symbols } = this.state;

    return (
      <div>
        {symbols.map((v) => (
          <Watch symbol={v} key={v} />
        ))}
      </div>
    );
  }
}
