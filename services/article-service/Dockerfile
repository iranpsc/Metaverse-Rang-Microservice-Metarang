FROM docker.arvancloud.ir/python:3.12.0-slim-bullseye

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

RUN rm -f /etc/apt/sources.list && \
    echo "deb http://mirror.mobinhost.com/debian/ bullseye main contrib non-free" > /etc/apt/sources.list && \
    echo "deb http://mirror.mobinhost.com/debian/ bullseye-updates main contrib non-free" >> /etc/apt/sources.list && \
    echo "deb http://mirror.mobinhost.com/debian-security bullseye-security main contrib non-free" >> /etc/apt/sources.list



RUN apt-get update && apt-get install -y \
    gcc \
    default-libmysqlclient-dev \
    pkg-config \
    build-essential \
    curl \
    netcat-openbsd \
    && rm -rf /var/lib/apt/lists/*


COPY requirements.txt /app/

RUN pip install -i https://package-mirror.liara.ir/repository/pypi/simple --no-cache-dir --upgrade --upgrade pip
RUN pip install -i https://package-mirror.liara.ir/repository/pypi/simple --no-cache-dir --upgrade -r requirements.txt

COPY ./core /app/

