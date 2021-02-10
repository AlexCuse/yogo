import React from "react";

// TODO: figure out how to use props in stateless func - can we skip prop type?
// eslint-disable-next-line react/prefer-stateless-function
export default class MarketOverview extends React.Component {
  constructor(props) {
    super(props);
    this.ref = React.createRef();
  }

  componentDidMount() {
    const script = document.createElement("script");
    script.src =
      "https://s3.tradingview.com/external-embedding/embed-widget-market-overview.js";
    script.async = true;
    script.innerHTML = JSON.stringify({
      colorTheme: "light",
      dateRange: "12M",
      showChart: true,
      locale: "en",
      largeChartUrl: "",
      isTransparent: false,
      showSymbolLogo: true,
      width: "800",
      height: "660",
      plotLineColorGrowing: "rgba(33, 150, 243, 1)",
      plotLineColorFalling: "rgba(33, 150, 243, 1)",
      gridLineColor: "rgba(240, 243, 250, 1)",
      scaleFontColor: "rgba(120, 123, 134, 1)",
      belowLineFillColorGrowing: "rgba(33, 150, 243, 0.12)",
      belowLineFillColorFalling: "rgba(33, 150, 243, 0.12)",
      symbolActiveColor: "rgba(33, 150, 243, 0.12)",
      tabs: [
        {
          title: "Indices",
          symbols: [
            {
              s: "FOREXCOM:SPXUSD",
              d: "S&P 500",
            },
            {
              s: "FOREXCOM:NSXUSD",
              d: "Nasdaq 100",
            },
            {
              s: "FOREXCOM:DJI",
              d: "Dow 30",
            },
            {
              s: "INDEX:NKY",
              d: "Nikkei 225",
            },
            {
              s: "INDEX:DEU30",
              d: "DAX Index",
            },
            {
              s: "FOREXCOM:UKXGBP",
              d: "FTSE 100",
            },
          ],
          originalTitle: "Indices",
        },
        {
          title: "Commodities",
          symbols: [
            {
              s: "CME_MINI:ES1!",
              d: "S&P 500",
            },
            {
              s: "CME:6E1!",
              d: "Euro",
            },
            {
              s: "COMEX:GC1!",
              d: "Gold",
            },
            {
              s: "NYMEX:CL1!",
              d: "Crude Oil",
            },
            {
              s: "NYMEX:NG1!",
              d: "Natural Gas",
            },
            {
              s: "CBOT:ZC1!",
              d: "Corn",
            },
          ],
          originalTitle: "Commodities",
        },
        {
          title: "Bonds",
          symbols: [
            {
              s: "CME:GE1!",
              d: "Eurodollar",
            },
            {
              s: "CBOT:ZB1!",
              d: "T-Bond",
            },
            {
              s: "CBOT:UB1!",
              d: "Ultra T-Bond",
            },
            {
              s: "EUREX:FGBL1!",
              d: "Euro Bund",
            },
            {
              s: "EUREX:FBTP1!",
              d: "Euro BTP",
            },
            {
              s: "EUREX:FGBM1!",
              d: "Euro BOBL",
            },
          ],
          originalTitle: "Bonds",
        },
        {
          title: "Forex",
          symbols: [
            {
              s: "FX:EURUSD",
            },
            {
              s: "FX:GBPUSD",
            },
            {
              s: "FX:USDJPY",
            },
            {
              s: "FX:USDCHF",
            },
            {
              s: "FX:AUDUSD",
            },
            {
              s: "FX:USDCAD",
            },
          ],
          originalTitle: "Forex",
        },
      ],
    });
    this.ref.current.appendChild(script);
  }

  render() {
    return (
      <div className="tradingview-widget-container" ref={this.ref}>
        <div className="tradingview-widget-container__widget" />
        <div className="tradingview-widget-copyright">
          Market Overview from TradingView
        </div>
      </div>
    );
  }
}
