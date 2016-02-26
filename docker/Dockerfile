FROM alpine:3.1
ENV DISUNITY_VERSION=v0.3.4
ENV UNITY3D2PNG_VERSION=v0.0.1
RUN apk add --update curl bash openjdk7-jre imagemagick
WORKDIR /root
RUN curl -L -O https://github.com/ata4/disunity/releases/download/${DISUNITY_VERSION}/disunity_${DISUNITY_VERSION}.zip && \
    unzip disunity_${DISUNITY_VERSION}.zip disunity.jar lib/*
RUN curl -L -O https://github.com/mix3/unity3d2png/releases/download/${UNITY3D2PNG_VERSION}/unity3d2png-${UNITY3D2PNG_VERSION}-linux-amd64.zip && \
    unzip unity3d2png-${UNITY3D2PNG_VERSION}-linux-amd64.zip unity3d2png/unity3d2png && \
    mv unity3d2png unity3d2png_dir && \
    mv unity3d2png_dir/unity3d2png unity3d2png && \
    rm -rf unity3d2png_dir/
CMD ["./unity3d2png"]
