FROM python:rc-alpine3.12 as builder

ENV PYTHONUNBUFFERED 1
ENV PATH="${PATH}:/sbin"
# Set build directory
WORKDIR /wheels

# Copy files necessary
COPY ./requirements.txt .

# Perform build and cleanup artifacts
RUN \
  apk add --no-cache \
  git \
  git-fast-import \
  && apk add --no-cache --virtual .build gcc musl-dev \
  && python -m pip install --upgrade pip \
  && pip install -r requirements.txt \
  && apk del .build gcc musl-dev \
  && rm -rf /usr/local/lib/python3.8/site-packages/mkdocs/themes/*/* \
  && rm -rf /tmp/*

# Set final MkDocs working directory
WORKDIR /docs

# Start development server by default
ENTRYPOINT ["mkdocs"]
CMD ["serve", "--dev-addr=0.0.0.0:8080"]
