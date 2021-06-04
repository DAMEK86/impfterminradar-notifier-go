FROM alpine:3

ARG UID=1001
ARG USER=app
ARG GID=1001
ARG GROUP=app
ENV WORKINGDIR /app

EXPOSE 8080

RUN apk --no-cache add ca-certificates

WORKDIR $WORKINGDIR
RUN addgroup -g $GID -S $GROUP && adduser -u $UID -S $USER -G $GROUP && \
    mkdir -p /app &&\
    chown -R $USER:$GROUP /app

COPY impftermin-notifier .
USER $USER
CMD ./impftermin-notifier