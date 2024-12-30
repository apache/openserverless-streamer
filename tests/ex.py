# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

import socket
import time

example_data = [
    "Hello, World!",
    "This is a test",
    "from an HTTP SSE request",
    "Through an openwhisk action",
    "To a socket server",
    "Back to the HTTP client"
]

def main(args):

    streamer = (args.get("STREAM_HOST"), args.get("STREAM_PORT"))

    if not streamer[0] or not streamer[1]:
        return {"body": "please provide a STREAM_HOST and STREAM_PORT"}

    print(f"streamer: {streamer}")

    # # invoke a call to a streaming api server like OpenAI
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.connect((streamer[0], int(streamer[1])))

        for ex in example_data:
            time.sleep(1)
            if ex:
                # send data
                print(f"sending: {ex}")
                s.sendall(ex.encode())

        print("done sending")
        s.close()

    return {"body": "done"}
