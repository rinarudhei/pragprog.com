FROM alpine:latest 
RUN apk add --no-cache ncurses
RUN mkdir /app && adduser -h /app -D pomo 
WORKDIR /app 
COPY --chown=pomo /pomo/pomo .
CMD ["/app/pomo"]
