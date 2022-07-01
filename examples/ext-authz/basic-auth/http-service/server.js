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
    res.writeHead(200, { "x-current-user": user }); // or `res.status(200);` or `res.writeHead(200);`

    res.end()
  } else {
    res.status(401)
    // res.status(403)

    res.end('Unauthorized - hint: credentials are kubeshop:kubeshop')
  }
})

app.listen(port, () => console.log(`kusk-ext-authz-http-service-basic-auth: server.js listening on port ${port}!`))
