<!--
  ~ Licensed to the Apache Software Foundation (ASF) under one
  ~ or more contributor license agreements.  See the NOTICE file
  ~ distributed with this work for additional information
  ~ regarding copyright ownership.  The ASF licenses this file
  ~ to you under the Apache License, Version 2.0 (the
  ~ "License"); you may not use this file except in compliance
  ~ with the License.  You may obtain a copy of the License at
  ~
  ~   http://www.apache.org/licenses/LICENSE-2.0
  ~
  ~ Unless required by applicable law or agreed to in writing,
  ~ software distributed under the License is distributed on an
  ~ "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
  ~ KIND, either express or implied.  See the License for the
  ~ specific language governing permissions and limitations
  ~ under the License.
-->

# Apache OpenServerless Streamer (incubating)

The OpenServerless streamer is a tool to relay a stream from OpenWhisk actions to an outside
HTTP client.

The streamer is a simple HTTP server that exposes an endpoint /stream/{namespace}/{action} to 
invoke the relative OpenWhisk action, open a socket for the action to write to, and relay the
output to the client.

It expects 2 environment variables to be set:
- `OW_APIHOST`: the OpenWhisk API host
- `STREAMER_ADDR`: the address of the streamer server for the OpenWhisk actions to connect to

Other environment variables can be set to configure the streamer:

- `HTTP_SERVER_PORT`: the port the streamer server listens on (default: 8181)


## Endpoints

The streamer exposes the following endpoints (use POST in case you need to send arguments to the action):

- `GET/POST /action/{namespace}/{action}`: to invoke the OpenWhisk action on the given namespace, default package, and action name. It requires an a Authorization header with Bearer token with the OpenWhisk AUTH token
- `GET/POST /action/{namespace}/{package}/{action}`: to invoke the OpenWhisk action on the given namespace, custom package, and action name. It requires an a Authorization header with Bearer token with the OpenWhisk AUTH token

- `GET/POST /web/{namespace}/{action}`: to invoke an OpenWhisk web action on the given namespace, default package, and action name.
- `GET/POST /web/{namespace}/{package}/{action}`: to invoke an OpenWhisk web action on the given namespace, custom package, and action name.
