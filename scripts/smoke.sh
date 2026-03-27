#!/usr/bin/env bash
set -euo pipefail

curl -i -H 'Host: public.bloop.to' http://127.0.0.1:28080/hello
curl -i -u gene:secretpass -H 'Host: basic.bloop.to' http://127.0.0.1:28080/hello
curl -i -H 'Host: token.bloop.to' -H 'X-Bloop-Token: topsecret' http://127.0.0.1:28080/hello
curl -i -X POST -H 'Host: public.bloop.to' --data 'ping=post-body' http://127.0.0.1:28080/submit
