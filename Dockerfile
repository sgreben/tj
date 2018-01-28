FROM scratch
COPY binaries/linux_x86_64/tj /tj
ENTRYPOINT [ "/tj" ]