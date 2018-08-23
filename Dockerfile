FROM peersafes/fabric-ccenv:1.0.4
MAINTAINER peersafe

ADD . /opt/gopath/src/github.com/b3log/wide
ADD vendor/ /opt/gopath/src/
RUN go install github.com/visualfc/gotools github.com/nsf/gocode github.com/bradfitz/goimports

RUN useradd wide && useradd runner

WORKDIR /opt/gopath/src/github.com/b3log/wide

EXPOSE 7070
