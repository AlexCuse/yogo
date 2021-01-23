import React from "react";
import "./App.css";
import Container from "@material-ui/core/Container";
import Typography from "@material-ui/core/Typography";
import Box from "@material-ui/core/Box";
import { MemoryRouter as Router } from "react-router";
import { Link as RouterLink, Route, Switch } from "react-router-dom";
import Link from "@material-ui/core/Link";
import List from "@material-ui/core/List";
import Watchlist from "./watch/Watchlist";
import Current from "./signals/Current";
import CurrentDetail from "./signals/CurrentDetail";

function App() {
  return (
    <Container>
      <Box my={4}>
        <Typography variant="h4" component="h1" gutterBottom>
          yogo ¯\_(ツ)_/¯
        </Typography>
      </Box>
      <Router>
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
          <Route
            exact
            path="/signals/:name/detail"
            component={CurrentDetail}
          />
          <Route exact path="/signals" component={Current} />
        </Switch>
      </Router>
    </Container>
  );
}

export default App;
