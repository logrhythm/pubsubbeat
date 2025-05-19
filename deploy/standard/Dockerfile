# Dockerfile to create a Beat image from any repository

#####################################################################
# Download Beat source code
FROM golang:1.13 as source
# Who is hosting the GitHub repo?
ARG BEAT_PUBLISHER 
# What is the name of the Beat?
ARG BEAT_NAME
# What is the Beat version/GitHub tag? Defaults to . (no tag)
ARG BEAT_VERSION=.
WORKDIR /go/src/github.com/$BEAT_PUBLISHER/$BEAT_NAME
# Clone the repo
RUN \
  SRC_DIR="/go/src/github.com/$BEAT_PUBLISHER/$BEAT_NAME" && \
  git clone "https://github.com/$BEAT_PUBLISHER/$BEAT_NAME" . 
#####################################################################

#####################################################################
# Build the Beat executeable 
FROM source as builder
# We must repeat ARGs here because ARG is per-phase
ARG BEAT_PUBLISHER
ARG BEAT_NAME
RUN apt-get update && apt-get upgrade -y && apt-get install openssl
WORKDIR /go/src/github.com/$BEAT_PUBLISHER/$BEAT_NAME
# Build the Beat
RUN \
  make
#####################################################################

#####################################################################
# Build the distroless image
FROM gcr.io/distroless/base-debian11 as runner

# Repeat ARGs again
ARG BEAT_PUBLISHER
ARG BEAT_NAME
RUN apt-get update && apt-get upgrade -y && apt-get install openssl
RUN apt-get update && apt-get upgrade -y && apt-get install libc6
# Copy files into distroless image
COPY --from=builder /go/src/github.com/$BEAT_PUBLISHER/$BEAT_NAME/$BEAT_NAME  /beat
COPY --from=builder /go/src/github.com/$BEAT_PUBLISHER/$BEAT_NAME/LICENSE  /
# Entrypoint is "beat" instead of "pubsubbeat" so this can be a template
ENTRYPOINT [ "/beat" ]
#####################################################################
