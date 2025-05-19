# Dockerfile to create a Beat image from any repository

#####################################################################
# Download Beat source code and determine version
FROM golang:1.13 as source

ARG BEAT_PUBLISHER
ARG BEAT_NAME

# Install git (curl/python3-yaml no longer needed)
RUN apt-get update && apt-get install -y git

# Set BEAT_VERSION as dev_ + current date (MMDDYY)
RUN export BEAT_VERSION=dev_$(date +%m%d%y) && echo $BEAT_VERSION > /beat_version.txt

# Set environment variable for later stages
ENV BEAT_VERSION_FILE=/beat_version.txt

WORKDIR /go/src/github.com/$BEAT_PUBLISHER/$BEAT_NAME

# Clone the repo (no version checkout since version is date-based)
RUN git clone "https://github.com/$BEAT_PUBLISHER/$BEAT_NAME" .

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