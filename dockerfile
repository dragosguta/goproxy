# iron/go:dev is the alpine image with the go tools added
FROM iron/go:dev
WORKDIR /app

ENV SRC=/go/src/github.com/dragosguta/goproxy

# Add the source code:
ADD . $SRC

# Build it:
RUN go get github.com/aws/aws-sdk-go/aws \
 github.com/dgrijalva/jwt-go \
 github.com/lestrrat-go/jwx/jwk

RUN cd $SRC; go build -o goproxy; cp goproxy /app/

EXPOSE 1330

ENTRYPOINT ["./goproxy"]