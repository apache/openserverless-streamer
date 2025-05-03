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

The OpenServerless streamer is a tool to relay a stream from OpenWhisk actions 
to an outside HTTP client.

The streamer is a simple HTTP server that exposes an endpoint 
/stream/{namespace}/{action} to invoke the relative OpenWhisk action, open a 
socket for the action to write to, and relay the output to the client.

It expects 2 environment variables to be set:
- `OW_APIHOST`: the OpenWhisk API host
- `STREAMER_ADDR`: the address of the streamer server for the OpenWhisk actions to connect to

Other environment variables can be set to configure the streamer:

- `HTTP_SERVER_PORT`: the port the streamer server listens on (default: 80)
  
Cors handling is handled through these variables:

- `CORS_ENABLED`: set to 1 or true to enable the CORS handler
- `CORS_ALLOW_ORIGIN`: this defaults to `*`
- `CORS_ALLOW_METHODS`: this defaults to `GET,POST,OPTIONS`
- `CORS_ALLOW_HEADERS`: this defaults to `Authorization,Content-Type`


## Endpoints

The streamer exposes the following endpoints (use POST in case you need to send 
arguments to the action):

- `GET/POST /action/{namespace}/{action}`: to invoke the OpenWhisk action on the 
given namespace, default package, and action name. It requires an Authorization
header with Bearer token with the OpenWhisk AUTH token.

- `GET/POST /action/{namespace}/{package}/{action}`: to invoke the OpenWhisk 
action on the given namespace, custom package, and action name. It requires an 
Authorization header with Bearer token with the OpenWhisk AUTH token

- `GET/POST /web/{namespace}/{action}`: to invoke an OpenWhisk web action on the 
given namespace, default package, and action name.

- `GET/POST /web/{namespace}/{package}/{action}`: to invoke an OpenWhisk web 
action on the given namespace, custom package, and action name.

## Tasks

Taskfile supports the following tasks:

```yaml
* build:              Build the streamer binary locally. This will create a binary named streamer in the current directory. 
* buildx:             Build the docker image using buildx. Set PUSH=1 to push the image to the registry.
* clean:              Clean up the build artifacts. This will remove the streamer binary and clean the go cache. 
* docker-login:       Login to the docker registry. Set REGISTRY=ghcr or REGISTRY=dockerhub in .env to use the respective registry. 
* image-tag:          Create a new tag for the current git commit.       
* run:                Run the streamer binary locally, using configuration from .env file 
* test:               Run the tests in the src directory. 
```

## Build and push

### Private registry or local image

To build an image and push it on a private repository, firstly choose which
registry you want to use.
Tasks support is for Github (ghcr) and Dockerhub (dockerhub).
So copy the `.env.example` to `.env` and configure the required variables for
authentication and set the `REGISTRY` and `NAMESPACE` accordly.

Now create a new tag

```bash
$ task image-tag
```
You should see an output like this:

```bash
Deleted tag '0.1.0-rc1-incubating.2504260711' (was 5a2f6d3)
0.1.0-rc1-incubating.2504260905
```

:bulb: **NOTE** If you leave unset `REGISTRY` a local `openserverless-streamer` 
image will be built, using the generated tag.

If you setup the `REGISTRY` and `NAMESPACE`, you can give a:

```bash
$ task docker-login
```

To build:

```bash
$ task buildx
```

To build and push

```bash
$ task buildx PUSH=1
```

### Apache repository
To build an official Apache OpensSrverless Streamer image, you
need to be a committer.

If you have the proper permissions, the build process will start pushing a
new tag to apache/openserverless-streamer repository.
So, for example,  if your tag is `0.1.0-rc1-incubating.2504260905` and your
git remote is `apache`

```bash
$ git push apache 0.1.0-rc1-incubating.2504260905
```

This will trigger the build workflow, and the process will be visible at
https://github.com/apache/openserverless-streamer/actions


