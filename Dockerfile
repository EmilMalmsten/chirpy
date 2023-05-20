FROM debian:stretch-slim

# COPY source destination
COPY chirpy /bin/chirpy

CMD ["/bin/chirpy"]