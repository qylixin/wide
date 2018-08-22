FROM peersafes/fabric-ccenv:1.0.4
MAINTAINER peersafe

ADD . /opt/go/src/github.com/b3log/wide
ADD vendor/ /opt/go/src/
RUN go install github.com/visualfc/gotools github.com/nsf/gocode github.com/bradfitz/goimports

RUN useradd wide && useradd runner

WORKDIR /opt/go/src/github.com/b3log/wide
RUN go build -v

EXPOSE 7070
