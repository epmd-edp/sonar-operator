FROM alpine:3.8

ENV OPERATOR=/usr/local/bin/sonar-operator \
    USER_UID=1001 \
    USER_NAME=sonar-operator \
    HOME=/home/sonar-operator

# install operator binary
COPY sonar-operator ${OPERATOR}

COPY bin /usr/local/bin
COPY configs /usr/local/configs

RUN  chmod u+x /usr/local/bin/user_setup && \
     chmod ugo+x /usr/local/bin/entrypoint && \
     /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
