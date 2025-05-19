# Dockerfile to create a Beat image from any repository

#####################################################################
# === DOWNLOAD SOURCE CODE ===
FROM golang:1.13 as source
ARG BEAT_PUBLISHER 
ARG BEAT_NAME
ARG BEAT_VERSION=.
WORKDIR /go/src/github.com/$BEAT_PUBLISHER/$BEAT_NAME

RUN git clone "https://github.com/$BEAT_PUBLISHER/$BEAT_NAME" .

#####################################################################
# === BUILD BEAT EXECUTABLE ===
FROM golang:1.13 as builder
ARG BEAT_PUBLISHER
ARG BEAT_NAME

WORKDIR /go/src/github.com/$BEAT_PUBLISHER/$BEAT_NAME
COPY --from=source /go/src/github.com/$BEAT_PUBLISHER/$BEAT_NAME /go/src/github.com/$BEAT_PUBLISHER/$BEAT_NAME

# Install required packages in build stage
RUN apt-get update && apt-get upgrade -y && apt-get install -y openssl make

RUN make

#####################################################################
# === CREATE FINAL DISTROLLESS IMAGE ===
FROM gcr.io/distroless/base-debian11 as runner
ARG BEAT_PUBLISHER
ARG BEAT_NAME

# Only copy built binary and license
COPY --from=builder /go/src/github.com/$BEAT_PUBLISHER/$BEAT_NAME/$BEAT_NAME /beat
COPY --from=builder /go/src/github.com/$BEAT_PUBLISHER/$BEAT_NAME/LICENSE /

ENTRYPOINT [ "/beat" ]
