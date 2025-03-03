FROM node:20-bookworm

########################################################################################################################
# All dependencies
########################################################################################################################

RUN apt-get update \
    && apt-get install --no-install-recommends --yes \
    coreutils \
    dnsutils \
    netcat-traditional \
    iputils-ping \
    iproute2 \
    wget \
    curl \
    python3 \
    vim \
    xfonts-75dpi \
    xfonts-base \
    wkhtmltopdf \
    libqt5core5a \
    libqt5gui5 \
    libqt5network5 \
    libqt5printsupport5 \
    libqt5svg5 \
    libqt5webkit5 \
    libqt5widgets5 \
    && rm -rf /var/lib/apt/lists/* \
    && apt install --autoremove

COPY wkhtnltopdf/wkhtmltopdf_0.12.6-1.bookworm_amd64.deb /tmp/

RUN dpkg -i /tmp/wkhtmltopdf_0.12.6-1.bookworm_amd64.deb && \
    apt-get install -f -y && \
    rm /tmp/wkhtmltopdf_0.12.6-1.bookworm_amd64.deb

########################################################################################################################
# cardjacker and save2pdf
########################################################################################################################
USER root

RUN groupadd -r chall && useradd -r -g chall -d /home/chall -s /bin/bash chall

RUN mkdir -p /home/chall/storage/

WORKDIR /home/chall/

COPY cardjacker ./cardjacker
COPY save2pdf ./save2pdf

RUN chown -R chall:chall /home/chall/

USER chall

RUN cd cardjacker && npm install
RUN cd save2pdf && npm install

########################################################################################################################
# Flag
########################################################################################################################
USER root

COPY bins/echo-fake-flag /usr/bin/echo-flag
RUN chown chall:chall /usr/bin/echo-flag && chmod 100 /usr/bin/echo-flag

########################################################################################################################
# Running cardjacker and save2pdf
########################################################################################################################


EXPOSE 2025

COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

CMD ["/entrypoint.sh"]
