import React from "react";
import "./App.css";
import Container from "@material-ui/core/Container";
import Typography from "@material-ui/core/Typography";
import Box from "@material-ui/core/Box";
import {
  Link as RouterLink,
  BrowserRouter as Router,
  Route,
  Switch,
} from "react-router-dom";
import Link from "@material-ui/core/Link";
import List from "@material-ui/core/List";
import Watchlist from "./watch/Watchlist";
import Current from "./signals/Current";
import CurrentDetail from "./signals/CurrentDetail";
import Detail from "./symbol/Detail";
import MarketOverview from "./tradingview/MarketOverview";

function App() {
  return (
    <Container>
      <Router>
        <Box my={4}>
          <Typography variant="h4" component="h1" gutterBottom>
            <Link component={RouterLink} to="/">
              yogo ¯\_(ツ)_/¯
            </Link>
          </Typography>
        </Box>
        <List>
          <Typography>
            <Link component={RouterLink} to="/watch">
              Watchlist
            </Link>
          </Typography>
          <Typography>
            <Link component={RouterLink} to="/signals">
              Signals
            </Link>
          </Typography>
        </List>
        <Switch>
          <Route exact path="/watch" component={Watchlist} />
          <Route exact path="/signals/:name/detail" component={CurrentDetail} />
          <Route exact path="/signals" component={Current} />
          <Route exact path="/symbol/:symbol" component={Detail} />
          <Route exact path="/" component={MarketOverview} />
        </Switch>
      </Router>
    </Container>
  );
}

export default App;
