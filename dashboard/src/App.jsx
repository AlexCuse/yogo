import React from "react";
import "./App.css";
import Container from "@material-ui/core/Container";
import Typography from "@material-ui/core/Typography";
import Box from "@material-ui/core/Box";
import Signallist from "./signals/Signallist";

function App() {
  return (
    <Container>
      <Box my={4}>
        <Typography variant="h4" component="h1" gutterBottom>
          yogo ¯\_(ツ)_/¯
        </Typography>
      </Box>

      <Signallist />
    </Container>
  );
}

export default App;
