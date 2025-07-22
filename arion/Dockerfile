FROM golang:1.19

WORKDIR /src
COPY . ./
RUN go mod download
RUN go build -o /arion

EXPOSE 4000
CMD ["/arion"]
