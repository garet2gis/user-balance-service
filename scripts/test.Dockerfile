FROM postgres:15-alpine
RUN apk add --no-cache git make musl-dev go busybox-suid build-base
# Set environment variables for the database
ENV POSTGRESUSER postgres
ENV DBUSER postgres
ENV DBNAME testdb
ENV DBHOST localhost
ENV DBHOST localhost
ENV DBPASS testdb
ENV POSTGRES_HOST_AUTH_METHOD trust

# Configure Go
ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin
# Give postgres user full access to Go path, needed to install dependencies
RUN chown -R postgres /go
# Create a workdir, this is needed to properly install Go dependencies
# chown on the postgres user so it can clone the repo here
WORKDIR testdir
RUN chown -R postgres /testdir
RUN mkdir /tmpdir
# used for local testing ONLY
COPY . .
RUN chown -R postgres /testdir/cmd/main
RUN chown -R postgres /testdir/internal
RUN chown -R postgres /testdir/pkg
RUN chown -R postgres /testdir/static/reports
# RUN find /tmpdir -name *.go -exec chown -R postgres {} \;

# create the entrypoint db script
RUN echo -e "\
#!/bin/bash \n \
git clone \${GIT_URL} ../testdir/tmpdir/ \n \
cp -R ../testdir/tmpdir/* ../testdir/ \n \
/bin/sh ../testdir/scripts/docker-test-commands.sh & \n \
COMMANDS_PID=\$! \n \
(while kill -0 \$COMMANDS_PID; do sleep 1; done) && pg_ctl stop & \
" >> /docker-entrypoint-initdb.d/entrypoint.sh
# chmod the entrypoint db script
RUN chmod +x /docker-entrypoint-initdb.d/entrypoint.sh