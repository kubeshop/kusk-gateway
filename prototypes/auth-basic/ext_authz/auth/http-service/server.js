// const Http = require("http");
// const path = require("path");

// const tokens = require(process.env.USERS ||
//   path.join(__dirname, "..", "users.json"));

// const server = new Http.Server((req, res) => {
//   console.log("request.headers=", req.headers);

//   const authorization = req.headers["authorization"] || "";
//   const extracted = authorization.split(" ");
//   console.log("extracted=", extracted);
//   if (extracted.length === 2 && extracted[0] === "Bearer") {
//     const user = checkToken(extracted[1]);
//     console.log("user=", user);
//     if (user !== undefined) {
//       // The authorization server returns a response with "x-current-user" header for a successful
//       // request.
//       res.writeHead(200, { "x-current-user": user });
//       return res.end();
//     }
//   }
//   res.writeHead(403);
//   res.end();
// });

// const port = process.env.PORT || 9002;
// server.listen(port);
// console.log(`starting HTTP server on: ${port}`);

// function checkToken(token) {
//   return tokens[token];
// }

const express = require('express')
const auth = require('basic-auth')
const app = express()
const port = process.env.PORT || 9002;

app.use((req, res, next) => {
  console.log("headers=", req.headers)
  console.log("auth(req)=", auth(req))

  const { name: user, pass } = auth(req)
  console.log("user=", user)
  console.log("pass=", pass)

  if (user === 'kubeshop' && pass === 'kubeshop') {
    // res.status(200)

    res.writeHead(200, { "x-current-user": user });
    // res.writeHead(200);

    res.end()
  } else {
    res.status(401)
    // res.status(403)
    res.end('Unauthorized - hint: credentials are kubeshop:kubeshop')
  }
})

app.listen(port, () => console.log(`auth.http-service: server.js listening on port ${port}!`))
