FROM rust:1.89-slim-bookworm AS planner
WORKDIR /app
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && printf 'fn main() {}\n' > src/main.rs && cargo build --release --locked && rm -rf src

FROM planner AS test
COPY src ./src
RUN cargo test --locked

FROM planner AS build
COPY src ./src
RUN touch src/main.rs && cargo build --release --locked

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl \
    && mkdir -p /etc/mum \
    && printf 'archivo de prueba Falco 201801391\n' > /etc/mum/sensitive-201801391.txt \
    && chmod 0444 /etc/mum/sensitive-201801391.txt \
    && rm -rf /var/lib/apt/lists/*
COPY --from=build /app/target/release/mum-rust-api /usr/local/bin/mum-rust-api
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/mum-rust-api"]
