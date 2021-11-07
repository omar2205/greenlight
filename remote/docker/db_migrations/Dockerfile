FROM ubuntu:20.04

RUN apt update && apt install curl tar -y

WORKDIR "/db_migrations"

# Download migrate tool
RUN curl -L --no-progress-meter https://github.com/golang-migrate/migrate/releases/download/v4.15.1/migrate.linux-amd64.tar.gz | tar xvz

RUN mv ./migrate /bin

COPY ./wait-for-it.sh /db_migrations/wait-for-it.sh

RUN chmod +x ./wait-for-it.sh

# Run `wait for it` and wait until db is online with no timeout
# then do the migration
CMD ["./wait-for-it.sh", "db:5432", "-t", "-s", "--", "migrate", "-path=/migrations/", "-database", "postgres://postgres:hello@db:5432/greenlight?sslmode=disable", "up"]