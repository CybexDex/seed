FROM ubuntu:16.04

RUN apt-get update && apt-get install -y vim expect

ENV PORT 8899

ARG VERSION
ARG BUILD_DATE

WORKDIR	/usr/app

VOLUME /usr/app

COPY . .

LABEL bn.version=$VERSION
LABEL bn.build_date=$BUILD_DATE
LABEL bn.build_cmd="docker build --force-rm --no-cache --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') --build-arg VERSION=$VERSION -t jadepool-seed:$VERSION -f ./Dockerfile ."
LABEL bn.run_cmd="docker run -d --name jadepool-seed -p 8899:8899 -v /data/seed-data:/usr/app jadepool-seed:$VERSION"

EXPOSE $PORT

CMD ./startup.sh $PORT
