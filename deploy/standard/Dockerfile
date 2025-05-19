# Dockerfile to create a Beat image from any repository

#####################################################################
# Download Beat source code and determine version
FROM golang:1.13 as source

ARG BEAT_PUBLISHER
ARG BEAT_NAME

# Install curl and python3-yaml for version extraction
RUN apt-get update && apt-get install -y curl python3-yaml git

# Fetch latest.yml and extract version for pubsubbeat, prefix with dev_
RUN curl -sSL https://raw.githubusercontent.com/logrhythm/versions/master/latest.yml -o /tmp/latest.yml && \
    export BEAT_VERSION=$(python3 -c "import yaml; d=yaml.safe_load(open('/tmp/latest.yml')); print('dev_' + d['pubsubbeat'])") && \
    echo $BEAT_VERSION > /beat_version.txt

# Set environment variable for later stages
ENV BEAT_VERSION_FILE=/beat_version.txt

WORKDIR /go/src/github.com/$BEAT_PUBLISHER/$BEAT_NAME

# Clone the repo and checkout the version (without dev_ prefix)
RUN \
  export RAW_VERSION=$(cat /beat_version.txt | sed 's/^dev_//') && \
  git clone "https://github.com/$BEAT_PUBLISHER/$BEAT_NAME" . && \
  git checkout $RAW_VERSION

#####################################################################
# Build the Beat executable 
FROM source as builder

ARG BEAT_PUBLISHER
ARG BEAT_NAME

RUN apt-get update && apt-get upgrade -y && apt-get install -y openssl make

WORKDIR /go/src/github.com/$BEAT_PUBLISHER/$BEAT_NAME

RUN make

#####################################################################
# Build the distroless image
FROM gcr.io/distroless/base-debian11 as runner

ARG BEAT_PUBLISHER
ARG BEAT_NAME

# Copy the version file from the builder
COPY --from=source /beat_version.txt /beat_version.txt

# Copy files into distroless image
COPY --from=builder /go/src/github.com/$BEAT_PUBLISHER/$BEAT_NAME/$BEAT_NAME  /beat
COPY --from=builder /go/src/github.com/$BEAT_PUBLISHER/$BEAT_NAME/LICENSE  /

# Set environment variable and label for version
ENV BEAT_VERSION_FILE=/beat_version.txt
ENV BEAT_VERSION=$(cat /beat_version.txt)
LABEL beat.version=$BEAT_VERSION

# Entrypoint is "beat" instead of "pubsubbeat" so this can be a template
ENTRYPOINT [ "/beat" ]
#####################################################################